package scene

import (
	gl "github.com/chsc/gogl/gl43"
	//"github.com/jteeuwen/glfw"
	//	"github.com/rwesterteiger/vectormath"
	"log"
	//"github.com/rwesterteiger/go-gltest/buffers"
	"github.com/rwesterteiger/go-gltest/geom"
	"github.com/rwesterteiger/go-gltest/shader"
	vmath "github.com/rwesterteiger/vectormath"
	"github.com/rwesterteiger/go-gltest/gbuffer"
	"github.com/rwesterteiger/go-gltest/buffers"
	"github.com/rwesterteiger/go-gltest/post"
	"github.com/rwesterteiger/go-gltest/lights"
)

const objVertexShaderSource = `
#version 430
layout (location = 0) in vec3 vtx;
layout (location = 1) in vec3 normal;

out vec4 vAmbient;
out vec3 vEyeSpaceNormal;
out vec4 vAlbedo;

layout (location = 0) uniform mat4 P;
layout (location = 4) uniform mat4 V;
layout (location = 8) uniform mat4 M;

layout (location = 12) uniform vec4 diffuseColor;
		
void main(void) {
	gl_Position = P * V * M * vec4(vtx,1);
	vEyeSpaceNormal = (V * M * vec4(normal, 0)).xyz;
	vAlbedo = diffuseColor;
}
`

const objFragShaderSource = `
#version 430

#define M_PI (3.14159265358979323846)

layout (location = 0) out vec4 fragAlbedo;
layout (location = 1) out vec3 fragNormal;

in vec3 vEyeSpaceNormal;
in vec4 vAlbedo;

void main(void)
{
	fragAlbedo = vAlbedo;
	fragNormal = vEyeSpaceNormal;
}
`


type Scene struct {
	w int
	h int

	camProjMat vmath.Matrix4 
	camViewMat vmath.Matrix4 

	objects []*geom.Object
	lights []lights.Light

	objShader *shader.Shader
	gbuf *gbuffer.GBuffer

	// scene is rendered into this for filtering
	outputFBO gl.Uint
	outputTex gl.Uint

	postFilters []post.PostProcessFilter

	fsQuadVAO *buffers.VAO
	blitShader *shader.Shader	
}

func makeColorFBO(w, h int) (fbo gl.Uint, colorTex gl.Uint) {
	gl.GenTextures(1, &colorTex)

	gl.BindTexture(gl.TEXTURE_2D, colorTex);
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, gl.Sizei(w), gl.Sizei(h), 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, colorTex, 0)
	
	drawBufs := []gl.Enum{ gl.COLOR_ATTACHMENT0 }
	gl.DrawBuffers(1, &(drawBufs[0]))
	
	result := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)

	if result != gl.FRAMEBUFFER_COMPLETE {
		log.Fatal("Error creating gbuffer FBO!")
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	return
}

func Make(w, h int) (s *Scene) {
	s = new(Scene)

	vmath.M4MakeIdentity(&s.camProjMat)
	vmath.M4MakeIdentity(&s.camViewMat)

	s.objShader = shader.Make()
	s.objShader.AddShaderSource(objVertexShaderSource, gl.VERTEX_SHADER)
	s.objShader.AddShaderSource(objFragShaderSource, gl.FRAGMENT_SHADER)
	s.objShader.Link()

	s.gbuf = gbuffer.Make(w,h)
	s.outputFBO, s.outputTex = makeColorFBO(w,h)
	s.fsQuadVAO = makeFullscreenQuadVAO()
	s.blitShader = makeBlitShader()
	return
}

func makeFullscreenQuadVAO() (*buffers.VAO) {
	vtxs := buffers.MakeVBOFromVec2s([]vmath.Vector2{ {-1, -1}, {1, -1}, {1, 1}, {-1, 1} })
	tcs := buffers.MakeVBOFromVec2s([]vmath.Vector2{ { 0,0 }, {1,0}, {1,1}, {0,1} })
	indices := []uint32{ 0, 1, 2, 2, 3, 0 }

	vao := buffers.MakeVAO(gl.TRIANGLES, 6)
	vao.AttachVBO(0, vtxs)
	vao.AttachVBO(1, tcs)
	vao.SetIndexBuffer(indices)

	return vao
}

func makeBlitShader() (s *shader.Shader) {
	const vSrc =`
	#version 430
	layout (location = 0) in vec2 vtx;
	layout (location = 1) in vec2 tc;

	out vec2 vTc;

	void main(void) {
		gl_Position = vec4(vtx.xy, 0, 1);
		vTc = tc;
	}
	`

	const fSrc = `
	#version 430
	layout (location = 0) out vec4 fragData;
	in vec2 vTc;

	layout (location = 0) uniform sampler2D inTex;

	void main(void)
	{
		fragData = texture2D(inTex, vTc);
	}
	`

	s = shader.Make()
	s.AddShaderSource(vSrc, gl.VERTEX_SHADER)
	s.AddShaderSource(fSrc, gl.FRAGMENT_SHADER)
	s.Link()

	return
}


func (s *Scene) Delete() {
	for _,o := range s.objects {
		o.Delete()
	}

	for _,l := range s.lights {
		l.Delete()
	}

	for _,f := range s.postFilters {
		f.Delete()
	}

	s.objShader.Delete()
	s.gbuf.Delete()
	s.fsQuadVAO.Delete()
	s.blitShader.Delete()
}

func (s *Scene) AddObject(obj *geom.Object) {
	s.objects = append(s.objects, obj)
}

func (s *Scene) AddLight(light lights.Light) {
	s.lights = append(s.lights, light)
}

func (s *Scene) AddPostFilter(f post.PostProcessFilter) {
	s.postFilters = append(s.postFilters, f)
}

func (s *Scene) SetCameraPerspective(fovyRadians, aspect, zNear, zFar float32) {
	vmath.M4MakePerspective(&s.camProjMat, fovyRadians, aspect, zNear, zFar)
}

func (s *Scene) SetCameraLookAt(eyePos, lookAtPos *vmath.Point3, upVec *vmath.Vector3) {
	vmath.M4MakeLookAt(&s.camViewMat, eyePos, lookAtPos, upVec)
}

func (s *Scene) doRender(P, V *vmath.Matrix4) {
	sh := s.objShader
	sh.ProgramUniformM4(0, P)
	sh.ProgramUniformM4(4, V)
	
	sh.BindFragDataLocation(0, "fragAlbedo")
	sh.BindFragDataLocation(1, "fragNormal")

	sh.Enable()

	for _, o := range s.objects {
		sh.ProgramUniformM4(8, o.GetModelMatrix())
		sh.ProgramUniformF4(12, o.GetDiffuseColor())
		o.Draw()
	}

	sh.Disable()
}

func (s *Scene) Render() {
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	gl.Enable(gl.DEPTH_TEST)

	for _, l := range s.lights {
		if l.NeedDepthPass() {
			projMat, viewMat := l.BeginDepthPass()
			s.doRender(projMat, viewMat)
			l.EndDepthPass()
		}
	}


	s.gbuf.Begin()
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	s.doRender(&s.camProjMat, &s.camViewMat)
	s.gbuf.End()

	gl.BindFramebuffer(gl.FRAMEBUFFER, s.outputFBO)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.ONE, gl.ONE)

	for _, l := range s.lights {
		l.Render(s.gbuf, &s.camProjMat, &s.camViewMat)
	}

	gl.Disable(gl.BLEND)

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	tex := s.outputTex

	for _, f := range s.postFilters {
		tex = f.Apply(s.gbuf, tex, &s.camProjMat, &s.camViewMat)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	s.blitShader.ProgramUniform1i(0,0)
	s.blitShader.Enable()
	s.fsQuadVAO.Draw()
	s.blitShader.Disable()
}
