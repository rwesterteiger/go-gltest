package gbuffer

import (
	gl "github.com/chsc/gogl/gl43"
	//vmath "github.com/rwesterteiger/vectormath"
	"log"
)

type GBuffer struct {
	w,h gl.Sizei
	fbo gl.Uint
	albedoTex gl.Uint
	normalTex gl.Uint
	depthTex gl.Uint
}

func setTexParameters() {
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
}

func Make(w,h int) (g *GBuffer) {
	g = new(GBuffer)
	
	g.w = gl.Sizei(w)
	g.h = gl.Sizei(h)

	gl.GenTextures(1, &g.albedoTex)
	gl.GenTextures(1, &g.normalTex)
	gl.GenTextures(1, &g.depthTex)

	gl.BindTexture(gl.TEXTURE_2D, g.albedoTex);
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, gl.Sizei(w), gl.Sizei(h), 0, gl.RGBA, gl.FLOAT, nil)
	setTexParameters()

	gl.BindTexture(gl.TEXTURE_2D, g.normalTex);
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB16F, gl.Sizei(w), gl.Sizei(h), 0, gl.RGB, gl.FLOAT, nil);
	setTexParameters()

	gl.BindTexture(gl.TEXTURE_2D, g.depthTex);
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT, gl.Sizei(w), gl.Sizei(h), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	setTexParameters()

	gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.GenFramebuffers(1, &g.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, g.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, g.albedoTex, 0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, g.normalTex, 0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, g.depthTex, 0)
	
	drawBufs := []gl.Enum{ gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1 }
	gl.DrawBuffers(2, &(drawBufs[0]))

	result := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)

	if result != gl.FRAMEBUFFER_COMPLETE {
		log.Fatal("Error creating gbuffer FBO!")
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	return
}

func (g *GBuffer) Delete() {
	gl.DeleteFramebuffers(1, &g.fbo)
	gl.DeleteTextures(1, &g.albedoTex)
	gl.DeleteTextures(1, &g.normalTex)
	gl.DeleteTextures(1, &g.depthTex)
}

func (g *GBuffer) Begin()  {
	gl.BindFramebuffer(gl.FRAMEBUFFER, g.fbo)
	gl.Viewport(0, 0, g.w, g.h)
}

func (_ *GBuffer) End() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (g *GBuffer) GetAlbedoTex() gl.Uint {
	return g.albedoTex
}

func (g *GBuffer) GetNormalTex() gl.Uint {
	return g.normalTex
}

func (g *GBuffer) GetDepthTex() gl.Uint {
	return g.depthTex
}

