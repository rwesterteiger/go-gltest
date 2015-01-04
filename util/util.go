package util

import (
	"github.com/go-gl/glow/gl-core/4.1/gl"
 	"log"
 	"math"
 	"fmt"
	vmath "github.com/rwesterteiger/vectormath"
)

type BBox struct {
	min vmath.Vector3
	max vmath.Vector3
}

func MakeBBox() (*BBox) {
	return &BBox {
		vmath.Vector3 { math.MaxFloat32, math.MaxFloat32, math.MaxFloat32 },
		vmath.Vector3 {-math.MaxFloat32,-math.MaxFloat32,-math.MaxFloat32 },
	}
}

func maxf(a float32, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}
func minf(a float32, b float32) float32 {
	return float32(math.Min(float64(a), float64(b)))
}

func (b *BBox) AddPoint(v *vmath.Vector3) {
	b.min.X = minf(b.min.X, v.X)
	b.min.Y = minf(b.min.Y, v.Y)
	b.min.Z = minf(b.min.Z, v.Z)

	b.max.X = maxf(b.max.X, v.X)
	b.max.Y = maxf(b.max.Y, v.Y)
	b.max.Z = maxf(b.max.Z, v.Z)
}

func (b *BBox) AddPoints(vtxs []vmath.Vector3) {
	for _,v := range vtxs {
		b.AddPoint(&v)
	}
}

func (b *BBox) GetCenter() vmath.Vector3 {
	var v vmath.Vector3

	vmath.V3Add(&v, &b.min, &b.max)
	vmath.V3ScalarMul(&v, &v, 0.5)

	return v
}
// implement Stringer interface
func (b *BBox) String() string {
	return fmt.Sprintf("[<%v, %v, %v> - <%v, %v, %v>]", b.min.X, b.min.Y, b.min.Z, b.max.X, b.max.Y, b.max.Z)
}

func BindTextures2D(names ...uint32) {
	for i,n := range names {
		gl.ActiveTexture(gl.TEXTURE0 + uint32(i))
		gl.BindTexture(gl.TEXTURE_2D, n)
	}
}

func CheckGL() {
	err := gl.GetError()

	if gl.GetError() != gl.NO_ERROR {
			log.Panicf("GL error: %#x", err)
	}
}

func MultiplyTransforms(result *vmath.Matrix4, xforms ...interface{}) {
	vmath.M4MakeIdentity(result)


	for _,x := range xforms {
		switch t := x.(type) {
		case *vmath.Matrix4:
			vmath.M4Mul(result, result, t)
		case *vmath.Transform3:
			vmath.M4MulT3(result, result, t)
		default:
			log.Panicf("Unexpected type %T in util.MultiplyTransforms()!", t)
		}
	}

	return
}