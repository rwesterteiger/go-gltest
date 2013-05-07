package lights

import (
	"github.com/rwesterteiger/go-gltest/shader"
	"github.com/rwesterteiger/go-gltest/buffers"
	"github.com/rwesterteiger/go-gltest/gbuffer"
	vmath "github.com/rwesterteiger/vectormath"
	gl "github.com/rwesterteiger/gogl/gl32"
)

const ambientLightVtxShaderSrc =`
	#version 430

	layout (location = 0) in vec3 vtx;

	layout (location = 1) noperspective out vec2 tc;
	void main(void) {
		gl_Position = vec4(vtx,1);
		tc = 0.5 * gl_Position.xy / gl_Position.w + 0.5;
	}
	`

const ambientLightFragShaderSrc = `
	#version 430

	#define M_PI (3.14159265358979323846)

	layout (location = 0) out vec4 fragData;
	layout (location = 1) noperspective in vec2 tc;

	layout (location = 0) uniform sampler2D albedoTex;

	void main(void)
	{
		fragData = 0.2 * texture2D(albedoTex, tc);
	}
`

type AmbientLight struct {
	shader *shader.Shader
	fsQuadVAO *buffers.VAO
}

func MakeAmbientLight() (s *AmbientLight) {
	s = new(AmbientLight)

	s.shader = shader.Make()
	s.shader.AddShaderSource(ambientLightVtxShaderSrc, gl.VERTEX_SHADER)
	s.shader.AddShaderSource(ambientLightFragShaderSrc, gl.FRAGMENT_SHADER)
	s.shader.Link()

	// make fullscreen quad VAO
	vtxs := buffers.MakeVBOFromVec2s([]vmath.Vector2{ {-1, -1}, {1, -1}, {1, 1}, {-1, 1} })
	tcs := buffers.MakeVBOFromVec2s([]vmath.Vector2{ { 0,0 }, {1,0}, {1,1}, {0,1} })
	indices := []uint32{ 0, 1, 2, 2, 3, 0 }

	s.fsQuadVAO = buffers.MakeVAO(gl.TRIANGLES, 6)
	s.fsQuadVAO.AttachVBO(0, vtxs)
	s.fsQuadVAO.AttachVBO(1, tcs)
	s.fsQuadVAO.SetIndexBuffer(indices)

	return
}

func (s *AmbientLight) Delete() {
	s.shader.Delete()
	s.fsQuadVAO.Delete()
}

func (_ *AmbientLight) NeedDepthPass() bool {
	return false
}

func (s *AmbientLight) BeginDepthPass() (projMat, viewMat *vmath.Matrix4) {
	return nil, nil
}

func (s *AmbientLight) EndDepthPass() {
}

func (s *AmbientLight) Render(gbuf *gbuffer.GBuffer, projMat, viewMat *vmath.Matrix4) {
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, gbuf.GetAlbedoTex())
	s.shader.ProgramUniform1i(0, 0)

	s.shader.Enable()
	s.fsQuadVAO.Draw()
	s.shader.Disable()

	gl.BindTexture(gl.TEXTURE_2D, 0)
}


