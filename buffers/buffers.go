package buffers

import (
	gl "github.com/rwesterteiger/gogl/gl32"
	vmath "github.com/rwesterteiger/vectormath"
)

type VBO struct {
	handle gl.Uint
	nComponents gl.Int
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

	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(4 * len(floatData)), gl.Pointer(&floatData[0]), gl.STATIC_DRAW)

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

	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(4 * len(floatData)), gl.Pointer(&floatData[0]), gl.STATIC_DRAW)

	return
}

func (vbo *VBO) Delete() {
	gl.DeleteBuffers(1, &vbo.handle)
}

func (vbo *VBO) vertexAttribPointer(idx gl.Uint) {
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.handle)
	defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.VertexAttribPointer(idx, vbo.nComponents, gl.FLOAT, gl.TRUE, 0, nil)	
}

type VAO struct {
	primitiveType gl.Enum
	nElements gl.Sizei
	handle gl.Uint
	idxBufferHandle gl.Uint
}

func MakeVAO(primitiveType gl.Enum, nElements int) (*VAO) {
	vao := &VAO{primitiveType : primitiveType, nElements : gl.Sizei(nElements) }

	gl.GenVertexArrays(1, &vao.handle)

	return vao
}
func (vao *VAO) Delete() {
	gl.DeleteBuffers(1, &vao.handle)
	gl.DeleteBuffers(1, &vao.idxBufferHandle)
}

func (vao *VAO) AttachVBO(vtxAttributeIdx gl.Uint, vbo *VBO) {
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
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(4 * len(indices)), gl.Pointer(&indices[0]), gl.STATIC_DRAW)
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
		gl.DrawElements(vao.primitiveType, vao.nElements, gl.UNSIGNED_INT, nil)
	} else {
		gl.DrawArrays(vao.primitiveType, 0, vao.nElements)
	}
}
