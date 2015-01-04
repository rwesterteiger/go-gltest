package post

import (
	"github.com/go-gl/glow/gl-core/4.1/gl"
	vmath "github.com/rwesterteiger/vectormath"
	//"log"
	"github.com/rwesterteiger/go-gltest/shader"
	//"github.com/rwesterteiger/go-gltest/buffers"
	"github.com/rwesterteiger/go-gltest/gbuffer"
	"github.com/rwesterteiger/go-gltest/util"
)

type dofShaderUniforms struct {
	InputTex 		int
	DepthTex 		int
	FocusDistance	float32
	InvP  			vmath.Matrix4
}

const dofVtxShaderSrc =`
	#version 330
	layout (location = 0) in vec2 vtx;
	noperspective out vec2 vTc;

	void main(void) {
		gl_Position = vec4(vtx.xy, 0, 1);
		vTc = 0.5 * gl_Position.xy / gl_Position.w + 0.5;
	}
	`

const dofFragShaderSrc = `
	#version 330
	layout (location = 0) out vec4 fragData;

	noperspective in vec2 vTc;

	uniform sampler2D InputTex;
	uniform sampler2D DepthTex;
	uniform float FocusDistance;
	uniform mat4 InvP;

	void main(void) {
		float z = texture(DepthTex, vTc).x;

		// determine eye-space position of pixel
    	vec4 vProjectedPos = 2 * vec4(vTc, z, 1.0) - 1;
		vec4 pos = InvP * vProjectedPos;  
		pos /= pos.w;
    
	
		float a = 40.0;
		float D = 0.03;
		float f = D * a;
		float cocRadius = abs(D * f * (FocusDistance - -pos.z) / (FocusDistance * (-pos.z - f)));

		/*
		// debug - mark zone of sharp focus
		if (cocRadius < 0.00002)  {
			fragData = 0.8 * vec4(1,1,1,1);
			return;
		}
		*/

		vec2 r = vec2(cocRadius);

		vec4 colorSum = vec4(0.0);
		colorSum += texture(InputTex, vTc + r * vec2(0.158509, -0.884836));
		colorSum += texture(InputTex, vTc + r * vec2(0.475528, -0.654508));
		colorSum += texture(InputTex, vTc + r * vec2(0.792547, -0.424181));
		colorSum += texture(InputTex, vTc + r * vec2(0.890511, -0.122678));
		colorSum += texture(InputTex, vTc + r * vec2(0.769421, 0.250000));
		colorSum += texture(InputTex, vTc + r * vec2(0.648330, 0.622678));
		colorSum += texture(InputTex, vTc + r * vec2(0.391857, 0.809017));
		colorSum += texture(InputTex, vTc + r * vec2(-0.000000, 0.809017));
		colorSum += texture(InputTex, vTc + r * vec2(-0.391857, 0.809017));
		colorSum += texture(InputTex, vTc + r * vec2(-0.648331, 0.622678));
		colorSum += texture(InputTex, vTc + r * vec2(-0.769421, 0.250000));
		colorSum += texture(InputTex, vTc + r * vec2(0.158509, -0.884836));
		colorSum += texture(InputTex, vTc + r * vec2(-0.890511, -0.122678));
		colorSum += texture(InputTex, vTc + r * vec2(-0.158509, -0.884836));
		colorSum += texture(InputTex, vTc + r * vec2(-0.475528, -0.654509));
		colorSum += texture(InputTex, vTc + r * vec2(-0.792547, -0.424181));
		colorSum += texture(InputTex, vTc + r * vec2(0.000000, -1.000000));
		colorSum += texture(InputTex, vTc + r * vec2(0.951056, -0.309017));
		colorSum += texture(InputTex, vTc + r * vec2(0.587785, 0.809017));
		colorSum += texture(InputTex, vTc + r * vec2(-0.587785, 0.809017));
		colorSum += texture(InputTex, vTc + r * vec2(-0.951057, -0.309017));
		colorSum += texture(InputTex, vTc + r * vec2(0.317019, -0.769672));
		colorSum += texture(InputTex, vTc + r * vec2(0.634038, -0.539345));
		colorSum += texture(InputTex, vTc + r * vec2(0.829966, 0.063661));
		colorSum += texture(InputTex, vTc + r * vec2(0.708876, 0.436339));
		colorSum += texture(InputTex, vTc + r * vec2(0.195928, 0.809017));
		colorSum += texture(InputTex, vTc + r * vec2(-0.195929, 0.809017));
		colorSum += texture(InputTex, vTc + r * vec2(-0.951057, -0.309017));
		colorSum += texture(InputTex, vTc + r * vec2(-0.708876, 0.436339));
		colorSum += texture(InputTex, vTc + r * vec2(-0.829966, 0.063661));
		colorSum += texture(InputTex, vTc + r * vec2(-0.317019, -0.769672));
		colorSum += texture(InputTex, vTc + r * vec2(-0.634038, -0.539345));
		colorSum += texture(InputTex, vTc + r * vec2(-0.951057, -0.309017));
		colorSum += texture(InputTex, vTc + r * vec2(-0.951057, -0.309017));


		fragData = colorSum / 34.0;
		
	}
	`

type DoFFilter struct {
	PostProcessFilterBase

	dofShader *shader.Shader
	focusDistance float32
}

func MakeDoFFilter(w, h int, focusDistance float32) (d *DoFFilter) {
	d = new(DoFFilter)
	d.PostProcessFilterBase.init(w,h)
	d.focusDistance = focusDistance

	d.dofShader = shader.Make()
	d.dofShader.AddShaderSource(dofVtxShaderSrc, gl.VERTEX_SHADER)
	d.dofShader.AddShaderSource(dofFragShaderSrc, gl.FRAGMENT_SHADER)
	d.dofShader.Link()

	return
}

func (b *DoFFilter) Delete() {
	b.PostProcessFilterBase.delete()
}

func (b *DoFFilter) Apply(gbuf *gbuffer.GBuffer, inputTex uint32, P, V *vmath.Matrix4) (outputTex uint32) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, b.outputFBO)
	
	util.BindTextures2D(inputTex, gbuf.GetDepthTex())

	//gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
	//gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
	//gl.GenerateMipmap(gl.TEXTURE_2D)

	uniforms := &dofShaderUniforms{ InputTex : 0, DepthTex : 1, FocusDistance : b.focusDistance }

	vmath.M4Inverse(&uniforms.InvP, P)

	b.dofShader.Enable()
	b.dofShader.SetUniforms(uniforms)
	b.fsQuadVAO.Draw() // downsample input texture into blurFBOs[0]
	b.dofShader.Disable()

	util.BindTextures2D(0, 0)

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	return b.outputTex
}



	
