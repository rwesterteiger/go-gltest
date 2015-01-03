package buffers

import (
	"github.com/go-gl/glow/gl-core/4.1/gl"
	vmath "github.com/rwesterteiger/vectormath"
	"fmt"
)

type VBO struct {
	handle uint32
	nComponents int32
}

func MakeVBOFromVec3s(vecs []vmath.Vector3) (vbo *VBO) {
	vbo = &VBO{ nComponents : 3 }
	gl.GenBuffers(1, &vbo.handle)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.handle)
	defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	floatData := make([]float32, 3 * len(vecs))
	for i := 0; i < len(vecs); i++ {
		floatData[3*i + 0] = vecs[i].X
		floatData[3*i + 1] = vecs[i].Y
		floatData[3*i + 2] = vecs[i].Z
	}

	gl.BufferData(gl.ARRAY_BUFFER, 4 * len(floatData), gl.Ptr(&floatData[0]), gl.STATIC_DRAW)

	return
}


func MakeVBOFromVec2s(vecs []vmath.Vector2) (vbo *VBO) {
	vbo = &VBO{ nComponents : 2 }
	gl.GenBuffers(1, &vbo.handle)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.handle)
	defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	floatData := make([]float32, 2 * len(vecs))
	for i := 0; i < len(vecs); i++ {
		floatData[2*i + 0] = vecs[i].X
		floatData[2*i + 1] = vecs[i].Y
	}

	gl.BufferData(gl.ARRAY_BUFFER, 4 * len(floatData), gl.Ptr(&floatData[0]), gl.STATIC_DRAW)

	return
}

func (vbo *VBO) Delete() {
	gl.DeleteBuffers(1, &vbo.handle)
}

func (vbo *VBO) vertexAttribPointer(idx uint32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.handle)
	defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.VertexAttribPointer(idx, vbo.nComponents, gl.FLOAT, true, 0, nil)	
}

type VAO struct {
	primitiveType uint32
	nElements int32
	handle uint32
	idxBufferHandle uint32
}

func MakeVAO(primitiveType uint32, nElements int) (*VAO) {
	vao := &VAO{primitiveType : primitiveType, nElements : int32(nElements) }

	gl.GenVertexArrays(1, &vao.handle)

	return vao
}
func (vao *VAO) Delete() {
	gl.DeleteBuffers(1, &vao.handle)
	gl.DeleteBuffers(1, &vao.idxBufferHandle)
}

func (vao *VAO) AttachVBO(vtxAttributeIdx uint32, vbo *VBO) {
	vao.bind()
	defer vao.unbind()
	
	vbo.vertexAttribPointer(vtxAttributeIdx)
	gl.EnableVertexAttribArray(vtxAttributeIdx)
}

func (vao *VAO) SetIndexBuffer(indices []uint32) {
	gl.DeleteBuffers(1, &vao.idxBufferHandle)
	gl.GenBuffers(1, &vao.idxBufferHandle)

	vao.bind()
	defer vao.unbind()

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, vao.idxBufferHandle)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4 * len(indices), gl.Ptr(&indices[0]), gl.STATIC_DRAW)
}
	
func (vao *VAO) bind() {
	gl.BindVertexArray(vao.handle)
}

func (_ *VAO) unbind() {
	gl.BindVertexArray(0)
}

func (vao *VAO) Draw() {
	vao.bind()
	defer vao.unbind()

	if vao.idxBufferHandle != 0 {
		if gl.GetError() != gl.NO_ERROR {
			panic("drawelements before")
		}
		gl.DrawElements(vao.primitiveType, vao.nElements, gl.UNSIGNED_INT, nil)

		err := gl.GetError()
		if err!= gl.NO_ERROR {
			fmt.Println(err)
			panic("drawelements after")
		}
	} else {

		if gl.GetError() != gl.NO_ERROR {
			panic("DrawArrays before")
		}
		gl.DrawArrays(vao.primitiveType, 0, vao.nElements)

		if gl.GetError() != gl.NO_ERROR {
			panic("DrawArrays after")
		}
	}
}
