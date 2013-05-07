package shadowmap

import (
	gl "github.com/rwesterteiger/gogl/gl32"
	//vmath "github.com/rwesterteiger/vectormath"
	"log"
)

type ShadowMap struct {
	shadowTex gl.Uint 
	fbo gl.Uint
}

func Make() (s *ShadowMap) {
	s = new(ShadowMap)
	gl.GenTextures(1, &s.shadowTex)

	gl.BindTexture(gl.TEXTURE_2D, s.shadowTex);
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT, 512, 512, 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
	borderColor := []gl.Float{ 0, 0, 0, 0 }
	gl.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR, &(borderColor[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_COMPARE_MODE, gl.COMPARE_REF_TO_TEXTURE);
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_COMPARE_FUNC, gl.LEQUAL);

	gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.GenFramebuffers(1, &s.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, s.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, s.shadowTex, 0)
	gl.DrawBuffer(gl.NONE)

	result := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)

	if result != gl.FRAMEBUFFER_COMPLETE {
		log.Fatal("Error creating shadowmap FBO!")
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	return
}

func (s *ShadowMap) Delete() {
	gl.DeleteTextures(1, &s.shadowTex)
	gl.DeleteFramebuffers(1, &s.fbo)
}

func (s *ShadowMap) GetDepthTex() gl.Uint {
	return s.shadowTex
}

func (s *ShadowMap) BeginDepthPass() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, s.fbo);
	gl.Viewport(0, 0, 512, 512)
	//gl.ClearDepth(0.0)
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	//gl.ClearDepth(1.0)
}

func (s *ShadowMap) EndDepthPass() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0);
}

