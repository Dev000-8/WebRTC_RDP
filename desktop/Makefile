
ifndef encoders
encoders = vp8
endif

tags = 
ifneq (,$(findstring h264,$(encoders)))
tags = h264enc
endif

ifneq (,$(findstring vp8,$(encoders)))
tags := $(tags) vp8enc
endif

tags := $(strip $(tags))

UNAME := $(shell uname 2>/dev/null || echo Unknown)
UNAME := $(patsubst CYGWIN%,Cygwin,$(UNAME))
UNAME := $(patsubst MSYS%,MSYS,$(UNAME))
UNAME := $(patsubst MINGW%,MSYS,$(UNAME))

all:
	go build -tags "$(tags)" -o ./product/agent.exe cmd/agent.go
	go build -o ./product/main.exe service/main.go

ifeq ($(UNAME),MSYS)
	cp -f /mingw64/bin/zlib1.dll ./product/zlib1.dll 
	cp -f /mingw64/bin/libvpx-1.dll ./product/libvpx-1.dll 
	cp -f /mingw64/bin/libwinpthread-1.dll ./product/libwinpthread-1.dll
endif
	
$(info ************  $(UNAME) ************)

.PHONY: clean
clean:
	@if [ -f ./product/agent ]; then rm ./product/agent; fi
	@if [ -f ./product/main ]; then rm ./product/main; fi
	@if [ -f ./product/zlib1.dll ]; then rm ./product/zlib1.dll; fi
	@if [ -f ./product/libvpx-1.dll ]; then rm ./product/libvpx-1.dll; fi
	@if [ -f ./product/libwinpthread-1.dll ]; then rm ./product/libwinpthread-1.dll; fi
