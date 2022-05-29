//go:build vp8enc
// +build vp8enc

package encoders

import (
	"bytes"
	"fmt"
	"image"
	"unsafe"
)

/*
#cgo pkg-config: vpx
#include <stdlib.h>
#include <string.h>
#include <vpx/vpx_encoder.h>
#include <vpx/vp8cx.h>

void rgba_to_yuv(uint8_t *destination, uint8_t *rgba, size_t width, size_t height, int32_t mode) {
	size_t image_size = width * height;
	size_t upos = image_size;
	size_t vpos = upos + upos / 4;
	size_t i = 0;

	for( size_t line = 0; line < height; ++line ) {
		if( !(line % 2) ) {
			for( size_t x = 0; x < width; x += 2 ) {
				uint8_t r = rgba[4 * i];
				uint8_t g = rgba[4 * i + 1];
				uint8_t b = rgba[4 * i + 2];
				uint8_t r0 = r;
				uint8_t g0 = g;
				uint8_t b0 = b;

				if(mode == 1)
				{
					r0 = r * 0.9;
					g0 = g * 0.9;
					b0 = b * 0.9;
				}
				else if(mode == 2)
				{
					r0 = r * 0.7;
					g0 = g * 0.7;
					b0 = b * 0.7;
				}
				else if(mode == 3)
				{
					r0 = r * 0.5;
					g0 = g * 0.5;
					b0 = b * 0.5;
				}

				destination[i++] = ((66*r0 + 129*g0 + 25*b0) >> 8) + 16;

				destination[upos++] = ((-38*r0 + -74*g0 + 112*b0) >> 8) + 128;
				destination[vpos++] = ((112*r0 + -94*g0 + -18*b0) >> 8) + 128;


				r = rgba[4 * i];
				g = rgba[4 * i + 1];
				b = rgba[4 * i + 2];

				r0 = r;
				g0 = g;
				b0 = b;

				if(mode == 1)
				{
					r0 = r * 0.9;
					g0 = g * 0.9;
					b0 = b * 0.9;
				}
				else if(mode == 2)
				{
					r0 = r * 0.7;
					g0 = g * 0.7;
					b0 = b * 0.7;
				}
				else if(mode == 3)
				{
					r0 = r * 0.5;
					g0 = g * 0.5;
					b0 = b * 0.5;
				}

				destination[i++] = ((66*r0 + 129*g0 + 25*b0) >> 8) + 16;
			}
		} else {
			for( size_t x = 0; x < width; x += 1 ) {
					uint8_t r = rgba[4 * i];
					uint8_t g = rgba[4 * i + 1];
					uint8_t b = rgba[4 * i + 2];

					uint8_t r0 = r;
					uint8_t g0 = g;
					uint8_t b0 = b;

					if(mode == 1)
					{
						r0 = r * 0.9;
						g0 = g * 0.9;
						b0 = b * 0.9;
					}
					else if(mode == 2)
					{
						r0 = r * 0.7;
						g0 = g * 0.7;
						b0 = b * 0.7;
					}
					else if(mode == 3)
					{
						r0 = r * 0.5;
						g0 = g * 0.5;
						b0 = b * 0.5;
					}

					destination[i++] = ((66*r0 + 129*g0 + 25*b0) >> 8) + 16;
			}
		}
	}
}

int vpx_img_plane_width(const vpx_image_t *img, int plane) {
  if (plane > 0 && img->x_chroma_shift > 0)
    return (img->d_w + 1) >> img->x_chroma_shift;
  else
    return img->d_w;
}

int vpx_img_plane_height(const vpx_image_t *img, int plane) {
  if (plane > 0 && img->y_chroma_shift > 0)
    return (img->d_h + 1) >> img->y_chroma_shift;
  else
    return img->d_h;
}

int vpx_img_read(vpx_image_t *img, void *bs) {
  int plane;
  for (plane = 0; plane < 3; ++plane) {
    unsigned char *buf = img->planes[plane];
    const int stride = img->stride[plane];
    const int w = vpx_img_plane_width(img, plane) *
                  ((img->fmt & VPX_IMG_FMT_HIGHBITDEPTH) ? 2 : 1);
    const int h = vpx_img_plane_height(img, plane);
    int y;
    for (y = 0; y < h; ++y) {
      memcpy(buf, bs, w);
      // if (fread(buf, 1, w, file) != (size_t)w) return 0;
      buf += stride;
      bs += w;
    }
  }
  return 1;
}

int32_t encode_frame(vpx_codec_ctx_t *ctx, vpx_image_t *img, int32_t framec, int32_t flags,
										 void *rgba, void *yuv_buf, int32_t w, int32_t h, void **encoded_frame, int32_t mode) {
	rgba_to_yuv(yuv_buf, rgba, w, h, mode);
	vpx_img_read(img, yuv_buf);
	if (vpx_codec_encode(ctx, img, (vpx_codec_pts_t)framec, 1, flags, VPX_DL_REALTIME) != 0) {
		return 0;
	}
	const vpx_codec_cx_pkt_t *pkt = NULL;
	vpx_codec_iter_t it = NULL;
	while ((pkt = vpx_codec_get_cx_data(ctx, &it)) != NULL) {
		if (pkt->kind == VPX_CODEC_CX_FRAME_PKT) {
			*encoded_frame = pkt->data.frame.buf;
			return pkt->data.frame.sz;
		}
	}
	*encoded_frame = (void *)0xDEADBEEF;
	return 0;
}

vpx_codec_err_t codec_enc_config_default(vpx_codec_enc_cfg_t *cfg) {
	return vpx_codec_enc_config_default(vpx_codec_vp8_cx(), cfg, 0);
}

vpx_codec_err_t codec_enc_init(vpx_codec_ctx_t *codec, vpx_codec_enc_cfg_t *cfg) {
	return vpx_codec_enc_init(codec, vpx_codec_vp8_cx(), cfg, 0);
}

*/
import "C"

const keyFrameInterval = 60

//VP8Encoder VP8 encoder
type VP8Encoder struct {
	buffer     *bytes.Buffer
	realSize   image.Point
	codecCtx   C.vpx_codec_ctx_t
	vpxImage   C.vpx_image_t
	yuvBuffer  []byte
	frameCount uint
	// vpxCodexIter C.vpx_codec_iter_t
}

func newVP8Encoder(size image.Point, frameRate int) (Encoder, error) {
	buffer := bytes.NewBuffer(make([]byte, 0))

	var cfg C.vpx_codec_enc_cfg_t
	if C.codec_enc_config_default(&cfg) != 0 {
		return nil, fmt.Errorf("Can't init default enc. config")
	}
	cfg.g_w = C.uint(size.X)
	cfg.g_h = C.uint(size.Y)
	cfg.g_timebase.num = 1
	cfg.g_timebase.den = C.int(frameRate)
	cfg.rc_target_bitrate = 90000
	cfg.g_error_resilient = 1

	var vpxCodecCtx C.vpx_codec_ctx_t
	if C.codec_enc_init(&vpxCodecCtx, &cfg) != 0 {
		return nil, fmt.Errorf("Failed to initialize enc ctx")
	}
	var vpxImage C.vpx_image_t
	if C.vpx_img_alloc(&vpxImage, C.VPX_IMG_FMT_I420, C.uint(size.X), C.uint(size.Y), 0) == nil {
		return nil, fmt.Errorf("Can't alloc. vpx image")
	}

	return &VP8Encoder{
		buffer:     buffer,
		realSize:   size,
		codecCtx:   vpxCodecCtx,
		vpxImage:   vpxImage,
		yuvBuffer:  make([]byte, size.X*size.Y*2),
		frameCount: 0,
	}, nil
}

var oldFrame *image.RGBA = nil

//Encode encodes a frame into a h264 payload
func (e *VP8Encoder) Encode(frame *image.RGBA, mode byte) ([]byte, error) {

	/*if oldFrame == nil {
		oldFrame = frame
	}

	if bytes.Compare(frame.Pix, oldFrame.Pix) == 0 {
		//e.frameCount = 0
		//return nil, nil
	} else {
		oldFrame = frame
	}*/

	encodedData := unsafe.Pointer(nil)
	var flags C.int
	if e.frameCount%keyFrameInterval == 0 {
		flags |= C.VPX_EFLAG_FORCE_KF
	}
	frameSize := C.encode_frame(
		&e.codecCtx,
		&e.vpxImage,
		C.int(e.frameCount),
		flags,
		unsafe.Pointer(&frame.Pix[0]),
		unsafe.Pointer(&e.yuvBuffer[0]),
		C.int(e.realSize.X),
		C.int(e.realSize.Y),
		&encodedData,
		C.int(mode),
	)
	e.frameCount++
	if int(frameSize) > 0 {

		encoded := C.GoBytes(encodedData, frameSize)
		return encoded, nil
		return nil, nil
	}
	return nil, nil
}

//Encode encodes a frame into a h264 payload
func (e *VP8Encoder) VideoSize() (image.Point, error) {
	return e.realSize, nil
}

//Close flushes and closes the inner x264 encoder
func (e *VP8Encoder) Close() error {
	C.vpx_img_free(&e.vpxImage)
	C.vpx_codec_destroy(&e.codecCtx)
	return nil
}

func init() {
	registeredEncoders[VP8Codec] = newVP8Encoder
}
