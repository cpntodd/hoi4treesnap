package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"testing"

	"github.com/xypwn/filediver/dds"
)

const (
	ddsHeaderFlagsCaps        = 0x1
	ddsHeaderFlagsHeight      = 0x2
	ddsHeaderFlagsWidth       = 0x4
	ddsHeaderFlagsPitch       = 0x8
	ddsHeaderFlagsPixelFormat = 0x1000
	ddsHeaderFlagsMipMapCount = 0x20000
	ddsHeaderFlagsLinearSize  = 0x80000

	ddsCapsTexture = 0x1000
	ddsCapsMipMap  = 0x400000
	ddsCapsComplex = 0x8

	ddsPfAlphaPixels = 0x1
	ddsPfFourCC      = 0x4
	ddsPfRGB         = 0x40

	dxgiFormatBC1UNorm     = 71
	d3d10ResourceTexture2D = 3
)

type ddsPixelFormat struct {
	Size        uint32
	Flags       uint32
	FourCC      [4]byte
	RGBBitCount uint32
	RBitMask    uint32
	GBitMask    uint32
	BBitMask    uint32
	ABitMask    uint32
}

type ddsHeader struct {
	Size              uint32
	Flags             uint32
	Height            uint32
	Width             uint32
	PitchOrLinearSize uint32
	Depth             uint32
	MipMapCount       uint32
	Reserved          [11]uint32
	PixelFormat       ddsPixelFormat
	Caps              uint32
	Caps2             uint32
	Caps3             uint32
	Caps4             uint32
	Reserved2         uint32
}

type ddsDX10Header struct {
	DXGIFormat        uint32
	ResourceDimension uint32
	MiscFlag          uint32
	ArraySize         uint32
	MiscFlags2        uint32
}

func TestDDSDecodingVariants(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		check   func(*testing.T, image.Image)
		checkEx func(*testing.T, []byte)
	}{
		{
			name: "dxt1_no_alpha_ignores_bad_linear_size",
			data: buildDDS(buildDXT1Header(4, 4, 0, 1, [4]byte{'D', 'X', 'T', '1'}), nil, buildDXT1Block(0xf800, 0x07e0, 0)),
			check: func(t *testing.T, img image.Image) {
				assertPixel(t, img, 0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
			},
		},
		{
			name: "dxt5_with_alpha",
			data: buildDDS(buildDXT1Header(4, 4, 1, 1, [4]byte{'D', 'X', 'T', '5'}), nil, buildDXT5Block(255, 0, 0x001f, 0x07e0, 0)),
			check: func(t *testing.T, img image.Image) {
				assertPixel(t, img, 0, 0, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
			},
		},
		{
			name: "argb8_uncompressed_bgra_reordered",
			data: buildDDS(buildARGB8Header(2, 2), nil,
				[]byte{0x33, 0x22, 0x11, 0x44, 0x03, 0x02, 0x01, 0xff},
				[]byte{0x66, 0x55, 0x44, 0x77, 0xcc, 0xbb, 0xaa, 0x88},
			),
			check: func(t *testing.T, img image.Image) {
				assertPixel(t, img, 0, 0, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0x44})
				assertPixel(t, img, 1, 1, color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0x88})
			},
		},
		{
			name: "dx10_bc1_with_mipmaps",
			data: buildDDS(
				buildDXT1Header(4, 4, 0, 2, [4]byte{'D', 'X', '1', '0'}),
				&ddsDX10Header{DXGIFormat: dxgiFormatBC1UNorm, ResourceDimension: d3d10ResourceTexture2D, ArraySize: 1},
				buildDXT1Block(0x07e0, 0xf800, 0),
				buildDXT1Block(0x001f, 0xf800, 0),
			),
			check: func(t *testing.T, img image.Image) {
				assertPixel(t, img, 0, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
			},
			checkEx: func(t *testing.T, data []byte) {
				tex, err := dds.Decode(bytes.NewReader(data), true)
				if err != nil {
					t.Fatalf("decode with mipmaps: %v", err)
				}
				if len(tex.Images) != 1 {
					t.Fatalf("expected 1 image, got %d", len(tex.Images))
				}
				if len(tex.Images[0].MipMaps) != 2 {
					t.Fatalf("expected 2 mipmaps, got %d", len(tex.Images[0].MipMaps))
				}
				assertPixel(t, tex.Images[0].MipMaps[1], 0, 0, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			img, format, err := image.Decode(bytes.NewReader(test.data))
			if err != nil {
				t.Fatalf("image.Decode failed: %v", err)
			}
			if format != "dds" {
				t.Fatalf("expected dds format, got %q", format)
			}
			test.check(t, img)
			if test.checkEx != nil {
				test.checkEx(t, test.data)
			}
		})
	}
}

func buildDDS(header ddsHeader, dx10 *ddsDX10Header, payloads ...[]byte) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 256))
	buf.WriteString("DDS ")
	mustBinaryWrite(buf, header)
	if dx10 != nil {
		mustBinaryWrite(buf, *dx10)
	}
	for _, payload := range payloads {
		buf.Write(payload)
	}
	return buf.Bytes()
}

func buildDXT1Header(width, height, linearSize, mipMapCount uint32, fourCC [4]byte) ddsHeader {
	flags := uint32(ddsHeaderFlagsCaps | ddsHeaderFlagsHeight | ddsHeaderFlagsWidth | ddsHeaderFlagsPixelFormat | ddsHeaderFlagsLinearSize)
	caps := uint32(ddsCapsTexture)
	if mipMapCount > 1 {
		flags |= ddsHeaderFlagsMipMapCount
		caps |= ddsCapsMipMap | ddsCapsComplex
	}
	return ddsHeader{
		Size:              124,
		Flags:             flags,
		Height:            height,
		Width:             width,
		PitchOrLinearSize: linearSize,
		MipMapCount:       mipMapCount,
		PixelFormat: ddsPixelFormat{
			Size:   32,
			Flags:  ddsPfFourCC,
			FourCC: fourCC,
		},
		Caps: caps,
	}
}

func buildARGB8Header(width, height uint32) ddsHeader {
	return ddsHeader{
		Size:              124,
		Flags:             ddsHeaderFlagsCaps | ddsHeaderFlagsHeight | ddsHeaderFlagsWidth | ddsHeaderFlagsPitch | ddsHeaderFlagsPixelFormat,
		Height:            height,
		Width:             width,
		PitchOrLinearSize: width * 4,
		MipMapCount:       1,
		PixelFormat: ddsPixelFormat{
			Size:        32,
			Flags:       ddsPfRGB | ddsPfAlphaPixels,
			RGBBitCount: 32,
			RBitMask:    0x00ff0000,
			GBitMask:    0x0000ff00,
			BBitMask:    0x000000ff,
			ABitMask:    0xff000000,
		},
		Caps: ddsCapsTexture,
	}
}

func buildDXT1Block(color0, color1 uint16, indices uint32) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	mustBinaryWrite(buf, color0)
	mustBinaryWrite(buf, color1)
	mustBinaryWrite(buf, indices)
	return buf.Bytes()
}

func buildDXT5Block(alpha0, alpha1 uint8, color0, color1 uint16, colorIndices uint32) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 16))
	buf.WriteByte(alpha0)
	buf.WriteByte(alpha1)
	buf.Write(make([]byte, 6))
	mustBinaryWrite(buf, color0)
	mustBinaryWrite(buf, color1)
	mustBinaryWrite(buf, colorIndices)
	return buf.Bytes()
}

func assertPixel(t *testing.T, img image.Image, x, y int, expected color.NRGBA) {
	t.Helper()
	actual := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
	if actual != expected {
		t.Fatalf("pixel at (%d,%d) = %#v, expected %#v", x, y, actual, expected)
	}
}

func mustBinaryWrite(buf *bytes.Buffer, value any) {
	if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
		panic(err)
	}
}