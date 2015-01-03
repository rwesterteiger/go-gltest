package lights

import (
	"github.com/rwesterteiger/go-gltest/shadowmap"
	"github.com/rwesterteiger/go-gltest/shader"
	"github.com/rwesterteiger/go-gltest/buffers"
	"github.com/rwesterteiger/go-gltest/gbuffer"
	vmath "github.com/rwesterteiger/vectormath"
	"math"
	"github.com/go-gl/glow/gl-core/4.1/gl"
	"fmt"
	
)

type spotLightShaderUniforms struct {
	PV 					vmath.Matrix4
	AlbedoTex 			int
	NormalTex 			int
	DepthTex 			int
	ShadowMapTex 		int
	LightPosAndAngle 	vmath.Vector4
	InvP 				vmath.Matrix4
	ShadowPV 			vmath.Matrix4
	Color 				vmath.Vector3
	LightDir 			vmath.Vector3
}


const dbgVtxShaderSrc = `
	#version 410

	layout (location = 0) in vec3 vtx;

	uniform mat4 PV;

	void main(void) {
		vec4 pos = PV * vec4(vtx,1);
		gl_Position = pos;
	}
`

const dbgFragShaderSrc = `
	#version 410
	
	layout (location = 0) out vec4 fragData;

	uniform sampler2D depthTex;

	void main(void)
	{
		ivec2 texCoord = ivec2(gl_FragCoord.xy);
		float z = texelFetch(depthTex, texCoord, 0).x;

		if (gl_FragCoord.z > z)
			discard;

		fragData = vec4(1,1,1,1);
	}
`


const spotLightVtxShaderSrc =`
	#version 410

	layout (location = 0) in vec3 vtx;
	// layout (location = 1) in vec2 tc;
	uniform mat4 PV;

	noperspective out vec2 tcNormalized;
	void main(void) {
		gl_Position = PV * vec4(vtx,1);
		tcNormalized = 0.5 * gl_Position.xy / gl_Position.w + 0.5;
	}
	`

const spotLightFragShaderSrc = `
	#version 410

	#define M_PI (3.14159265358979323846)

	layout (location = 0) out vec4 fragData;
	noperspective in vec2 tcNormalized;

	uniform sampler2D AlbedoTex;
	uniform sampler2D NormalTex;
	uniform sampler2D DepthTex;
	uniform sampler2DShadow ShadowMapTex;
	uniform vec4 LightPosAndAngle; // xyz = eyespace pos, w = opening angle
	uniform mat4 InvP; // camera NDC -> viewspace
	uniform mat4 ShadowPV; // camera viewspace --> shadow clipspace
	uniform vec3 Color;

	uniform vec3 LightDir;

	const mat4 bias = mat4(	0.5, 0.0, 0.0, 0.0,
		               		0.0, 0.5, 0.0, 0.0,
			       			0.0, 0.0, 0.5, 0.0,
        	               	0.5, 0.5, 0.5, 1.0); 

	float getShadowAttenuation(vec3 pos) {
		vec4 vShadowCoord = bias * ShadowPV * vec4(pos, 1);
		vShadowCoord.z -= 0.02;

		float d = 0.0;

		d += textureProjOffset(ShadowMapTex, vShadowCoord, ivec2(-1,-1));
		d += textureProjOffset(ShadowMapTex, vShadowCoord, ivec2( 0,-1));
		d += textureProjOffset(ShadowMapTex, vShadowCoord, ivec2( 1,-1));

		d += textureProjOffset(ShadowMapTex, vShadowCoord, ivec2(-1, 0));
		d += textureProjOffset(ShadowMapTex, vShadowCoord, ivec2( 1, 0));

		d += textureProjOffset(ShadowMapTex, vShadowCoord, ivec2(-1, 1));
		d += textureProjOffset(ShadowMapTex, vShadowCoord, ivec2( 0, 1));
		d += textureProjOffset(ShadowMapTex, vShadowCoord, ivec2( 1, 1));

		d /= 8.0;

		return d;

	}

	void main(void)
	{
		vec4 diffuseMaterial = texture(AlbedoTex, tcNormalized);
		float z = texture(DepthTex, tcNormalized).x;
		vec3  n = texture(NormalTex, tcNormalized).xyz; // eyespace normal

		// determine eye-space position of pixel
    	vec4 vProjectedPos = 2 * vec4(tcNormalized, z, 1.0) - 1;
		vec4 pos = InvP * vProjectedPos;  
		pos /= pos.w;
	
		vec3 L = normalize(pos.xyz - LightPosAndAngle.xyz);
		float NdotL = -dot(L, n);

		if (NdotL < 0.0) {
			discard;
		}

		float angle = acos(dot(LightDir, L));
		float attenuation = 1 - smoothstep(LightPosAndAngle.w * 0.7, LightPosAndAngle.w * 0.8, angle);

		vec4 diffuse = vec4(Color, 1) * diffuseMaterial * max(0.0, NdotL);

		float specStrength = max(0.0, dot(reflect(-LightDir, n), normalize(pos.xyz)));
		vec4 specular = vec4(1 * pow(specStrength, 16));
		fragData = getShadowAttenuation(pos.xyz) * attenuation * (diffuse + specular);
	}
`


type SpotLight struct {
	pos vmath.Point3
	projMat vmath.Matrix4
	viewMat vmath.Matrix4
	pvMat vmath.Matrix4
	alpha float32

	shadowMap *shadowmap.ShadowMap
	shader *shader.Shader
	dbgShader *shader.Shader

	fsQuadVAO *buffers.VAO

	coneVAO *buffers.VAO

	color vmath.Vector3

	dir vmath.Vector3
}

func MakeSpotLight(pos, lookAt *vmath.Point3, up *vmath.Vector3, sceneBoundingSphereRadius float32, color *vmath.Vector3) (s *SpotLight) {
	s = new(SpotLight)

	vmath.P3Copy(&s.pos, pos)

	var diff vmath.Vector3
	vmath.P3Sub(&diff, lookAt, pos)
	s.alpha = float32(math.Asin(float64(sceneBoundingSphereRadius / diff.Length()))) // spotlight opening angle

	vmath.M4MakePerspective(&s.projMat, 2*s.alpha, 1.0, 1.0, 100.0)
	vmath.M4MakeLookAt(&s.viewMat, pos, lookAt, up)
	vmath.M4Mul(&s.pvMat, &s.projMat, &s.viewMat)

	vmath.P3Sub(&s.dir, lookAt, pos)
	vmath.V3Normalize(&s.dir, &s.dir)

	s.shadowMap = shadowmap.Make()


	// make shader to render light contribution into light accumulation buffer
	s.shader = shader.Make()
	s.shader.AddShaderSource(spotLightVtxShaderSrc, gl.VERTEX_SHADER)
	s.shader.AddShaderSource(spotLightFragShaderSrc, gl.FRAGMENT_SHADER)
	s.shader.Link()

	s.dbgShader = shader.Make()
	s.dbgShader.AddShaderSource(dbgVtxShaderSrc, gl.VERTEX_SHADER)
	s.dbgShader.AddShaderSource(dbgFragShaderSrc, gl.FRAGMENT_SHADER)
	s.dbgShader.Link()

	// make fullscreen quad VAO
	vtxs := buffers.MakeVBOFromVec2s([]vmath.Vector2{ {-1, -1}, {1, -1}, {1, 1}, {-1, 1} })
	tcs := buffers.MakeVBOFromVec2s([]vmath.Vector2{ { 0,0 }, {1,0}, {1,1}, {0,1} })
	indices := []uint32{ 0, 1, 2, 2, 3, 0 }

	s.fsQuadVAO = buffers.MakeVAO(gl.TRIANGLES, 6)
	s.fsQuadVAO.AttachVBO(0, vtxs)
	s.fsQuadVAO.AttachVBO(1, tcs)
	s.fsQuadVAO.SetIndexBuffer(indices)

	s.makeConeVAO()

	vmath.V3Copy(&s.color, color)
	return
}

func (s* SpotLight) makeConeVAO() {
	const vtxCount = 16
	// make cone VAO in eye-space
	vtxs := make([]vmath.Vector3, vtxCount+1)
	indices := make([]uint32, 3*vtxCount)

	coneLength := 10.0
	coneBaseRadius := coneLength * math.Sin(float64(s.alpha))

	for i := 0; i < vtxCount; i++ {
		beta := float64(i) * 2 * math.Pi / (vtxCount-1)
		x := float32(coneBaseRadius * math.Sin(beta))
		y := float32(coneBaseRadius * math.Cos(beta))
		vtxs[i] = vmath.Vector3{x, y, -float32(coneLength) }
		fmt.Printf("%v\n", vtxs[i])
	}
	// tip vtx of cone
	vtxs[vtxCount] = vmath.Vector3{0,0,0}

	// transform vertices to world space using inverse view matrix
	var invV vmath.Matrix4
	vmath.M4Inverse(&invV, &s.viewMat)

	for i := 0; i < vtxCount+1; i++ {
		var wsVtx vmath.Vector4;

		vmath.V4MakeFromV3(&wsVtx, &vtxs[i])
		wsVtx.W = 1.0
		vmath.M4MulV4(&wsVtx, &invV, &wsVtx)

		vtxs[i].X = wsVtx.X
		vtxs[i].Y = wsVtx.Y
		vtxs[i].Z = wsVtx.Z
	}
			
	for i := 0; i < vtxCount; i++ {
		indices[3*i+0] = uint32(i)
		indices[3*i+1] = uint32(vtxCount)
		indices[3*i+2] = uint32((i+1) % vtxCount)
	}

	
	s.coneVAO = buffers.MakeVAO(gl.TRIANGLES, len(indices))
	s.coneVAO.AttachVBO(0, buffers.MakeVBOFromVec3s(vtxs))
	//s.fsQuadVAO.AttachVBO(1, tcs)
	s.coneVAO.SetIndexBuffer(indices)
}

func (s *SpotLight) Delete() {
	s.shadowMap.Delete()
	s.shader.Delete()
	s.fsQuadVAO.Delete()
}

func (_ *SpotLight) NeedDepthPass() bool {
	return true
}

func (s *SpotLight) BeginDepthPass() (projMat, viewMat *vmath.Matrix4) {
	s.shadowMap.BeginDepthPass()
	return &s.projMat, &s.viewMat
}

func (s *SpotLight) EndDepthPass() {
	s.shadowMap.EndDepthPass()
}


func (s *SpotLight) Render(gbuf *gbuffer.GBuffer, projMat, viewMat *vmath.Matrix4) {

	if gl.GetError() != gl.NO_ERROR {
			panic("gl error in spotlight begin render")
	}

	uniforms := &spotLightShaderUniforms{
		AlbedoTex    : 0,
		NormalTex    : 1,
		DepthTex     : 2,
		ShadowMapTex : 3,
		Color        : s.color,
	}

	vmath.M4Inverse(&uniforms.InvP, projMat)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, gbuf.GetAlbedoTex())

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, gbuf.GetNormalTex())

	gl.ActiveTexture(gl.TEXTURE2)
	gl.BindTexture(gl.TEXTURE_2D, gbuf.GetDepthTex())

	gl.ActiveTexture(gl.TEXTURE3)
	gl.BindTexture(gl.TEXTURE_2D, s.shadowMap.GetDepthTex())

	var eyeSpacePos vmath.Vector4
	vmath.V4MakeFromP3(&eyeSpacePos, &s.pos)
	vmath.M4MulV4(&uniforms.LightPosAndAngle, viewMat, &eyeSpacePos)

	vmath.M4Mul(&uniforms.PV, projMat, viewMat)



	var invV vmath.Matrix4
	vmath.M4Inverse(&invV, viewMat)
	vmath.M4Mul(&uniforms.ShadowPV, &s.pvMat, &invV)

	var tmp vmath.Vector4
	vmath.M4MulV3(&tmp, viewMat, &s.dir) // eye space direction of light
	vmath.V3MakeFromElems(&uniforms.LightDir, tmp.X, tmp.Y, tmp.Z)

	if gl.GetError() != gl.NO_ERROR {
			panic("gl error in spotlight end render")
	}
	s.shader.Enable()
	s.shader.SetUniforms(uniforms)

	//s.fsQuadVAO.Draw()
	s.coneVAO.Draw()
	s.shader.Disable()


	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.ActiveTexture(gl.TEXTURE2)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
/*

	gl.LineWidth(2.0)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, gbuf.GetDepthTex())
	
	s.dbgShader.ProgramUniformM4(4, &PV)
	s.dbgShader.ProgramUniform1i(0, 0)


	s.dbgShader.Enable()
	s.coneVAO.Draw()
	s.dbgShader.Disable()

	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
*/



}


