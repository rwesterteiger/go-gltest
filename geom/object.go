package geom

import (
	"github.com/rwesterteiger/go-gltest/buffers"
	"github.com/rwesterteiger/go-gltest/util"
	vmath "github.com/rwesterteiger/vectormath"
)

type Object struct {
	vao *buffers.VAO
	diffuseColor vmath.Vector4
	modelMat vmath.Matrix4
	bbox util.BBox
}

func MakeObject(vao *buffers.VAO, bbox *util.BBox, diffuseColor *vmath.Vector4) (o *Object) {
	o = &Object{ vao : vao, diffuseColor : *diffuseColor, bbox : *bbox }

	// vmath.V4Copy(&o.diffuseColor, diffuseColor)
	vmath.M4MakeIdentity(&o.modelMat)

	return
}

func (o *Object) Delete() {
	o.vao.Delete()
}

func (o *Object) Draw() {
	o.vao.Draw()
}

func (o *Object) GetDiffuseColor() (*vmath.Vector4) {
	return &o.diffuseColor
}

func (o *Object) SetModelMatrix(M *vmath.Matrix4) {
	vmath.M4Copy(&o.modelMat, M)
}

func (o *Object) GetModelMatrix() (*vmath.Matrix4) {
	return &o.modelMat
}

func (o *Object) GetBoundingBox() (*util.BBox) {
	return &o.bbox
}