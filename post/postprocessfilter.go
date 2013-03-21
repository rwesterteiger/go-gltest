package post

import (
	gl "github.com/chsc/gogl/gl43"
	"github.com/rwesterteiger/go-gltest/gbuffer"
	"github.com/rwesterteiger/go-gltest/buffers"
	vmath "github.com/rwesterteiger/vectormath"
	"log"
)

type PostProcessFilter interface {
	Apply(gbuf *gbuffer.GBuffer, inputTex gl.Uint, P, V *vmath.Matrix4) (outputTex gl.Uint)
	Delete()
}

type PostProcessFilterBase struct {
	w, h int

	outputFBO gl.Uint
	outputTex gl.Uint
	fsQuadVAO *buffers.VAO
}

func makeFullscreenQuadVAO() (*buffers.VAO) {
	vtxs := buffers.MakeVBOFromVec2s([]vmath.Vector2{ {-1, -1}, {1, -1}, {1, 1}, {-1, 1} })
	indices := []uint32{ 0, 1, 2, 2, 3, 0 }

	vao := buffers.MakeVAO(gl.TRIANGLES, 6)
	vao.AttachVBO(0, vtxs)
	vao.SetIndexBuffer(indices)

	return vao
}


func setTexParameters() {
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
}

func checkFramebuffer() {
	result := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)

	if result != gl.FRAMEBUFFER_COMPLETE {
		log.Fatal("Error creating gbuffer FBO!")
	}
}

func (f *PostProcessFilterBase) init(w, h int) {
	f.w = w
	f.h = h

	// input FBO
	gl.GenTextures(1, &f.outputTex)

	gl.BindTexture(gl.TEXTURE_2D, f.outputTex);
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, gl.Sizei(w), gl.Sizei(h), 0, gl.RGBA, gl.FLOAT, nil)
	setTexParameters()

	gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.GenFramebuffers(1, &f.outputFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, f.outputFBO)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, f.outputTex, 0)
	
	drawBufs := []gl.Enum{ gl.COLOR_ATTACHMENT0 }
	gl.DrawBuffers(1, &(drawBufs[0]))

	checkFramebuffer()

	f.fsQuadVAO = makeFullscreenQuadVAO()
}

func (f *PostProcessFilterBase) delete() {
	gl.DeleteFramebuffers(1, &f.outputFBO)
	gl.DeleteTextures(1, &f.outputTex)
}
