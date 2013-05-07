package post

import (
	gl "github.com/rwesterteiger/gogl/gl32"
	vmath "github.com/rwesterteiger/vectormath"
	//"log"
	"github.com/rwesterteiger/go-gltest/shader"
	//"github.com/rwesterteiger/go-gltest/buffers"
	"github.com/rwesterteiger/go-gltest/gbuffer"
)

const vtxShaderSrc =`
	#version 430
	layout (location = 0) in vec2 vtx;
	noperspective out vec2 vTc;

	void main(void) {
		gl_Position = vec4(vtx.xy, 0, 1);
		vTc = 0.5 * gl_Position.xy / gl_Position.w + 0.5;
	}
	`

const downSampleFragShaderSrc = `
	#version 430
	layout (location = 0) out vec4 fragData;

	layout (location = 0) uniform sampler2D inputTex;
	layout (location = 1) uniform vec2 inputTexelSize;

	void main(void) {
		vec2 tc = (2 + 4 * gl_FragCoord.xy) * inputTexelSize;
		vec4 a = texture2D(inputTex, tc + vec2(-1,-1) * inputTexelSize);
		vec4 b = texture2D(inputTex, tc + vec2( 1,-1) * inputTexelSize);
		vec4 c = texture2D(inputTex, tc + vec2( 1, 1) * inputTexelSize);
		vec4 d = texture2D(inputTex, tc + vec2(-1, 1) * inputTexelSize);
		
	
		fragData = (a+b+c+d) / 4.0;
		fragData = fragData * fragData; // brightpass
	}
	`

const blurXFragShaderSrc = `
	#version 430
	layout (location = 0) out vec4 fragData;

	layout (location = 0) uniform sampler2D inputTex;
	layout (location = 1) uniform vec2 inputTexelSize;

	void main(void) {
		vec2 tc = gl_FragCoord.xy * inputTexelSize;
		vec4 sum = vec4(0);
	
		sum += 0.0162162162 * texture2D(inputTex, vec2(tc.x - 4 * inputTexelSize.x, tc.y));
		sum += 0.0540540541 * texture2D(inputTex, vec2(tc.x - 3 * inputTexelSize.x, tc.y));
		sum += 0.1216216216 * texture2D(inputTex, vec2(tc.x - 2 * inputTexelSize.x, tc.y));
		sum += 0.1945945946 * texture2D(inputTex, vec2(tc.x - 1 * inputTexelSize.x, tc.y));
		sum += 0.2270270270 * texture2D(inputTex, vec2(tc.x - 0 * inputTexelSize.x, tc.y));
		sum += 0.1945945946 * texture2D(inputTex, vec2(tc.x + 1 * inputTexelSize.x, tc.y));
		sum += 0.1216216216 * texture2D(inputTex, vec2(tc.x + 2 * inputTexelSize.x, tc.y));
		sum += 0.0540540541 * texture2D(inputTex, vec2(tc.x + 3 * inputTexelSize.x, tc.y));
		sum += 0.0162162162 * texture2D(inputTex, vec2(tc.x + 4 * inputTexelSize.x, tc.y));
		
		fragData = sum;
	}
	`


const blurYFragShaderSrc = `
	#version 430
	layout (location = 0) out vec4 fragData;

	layout (location = 0) uniform sampler2D inputTex;
	layout (location = 1) uniform vec2 inputTexelSize;

	void main(void) {
		vec2 tc = gl_FragCoord.xy * inputTexelSize;
		vec4 sum = vec4(0);
	
		sum += 0.0162162162 * texture2D(inputTex, vec2(tc.x, tc.y - 4 * inputTexelSize.y));
		sum += 0.0540540541 * texture2D(inputTex, vec2(tc.x, tc.y - 3 * inputTexelSize.y));
		sum += 0.1216216216 * texture2D(inputTex, vec2(tc.x, tc.y - 2 * inputTexelSize.y));
		sum += 0.1945945946 * texture2D(inputTex, vec2(tc.x, tc.y - 1 * inputTexelSize.y));
		sum += 0.2270270270 * texture2D(inputTex, vec2(tc.x, tc.y - 0 * inputTexelSize.y));
		sum += 0.1945945946 * texture2D(inputTex, vec2(tc.x, tc.y + 1 * inputTexelSize.y));
		sum += 0.1216216216 * texture2D(inputTex, vec2(tc.x, tc.y + 2 * inputTexelSize.y));
		sum += 0.0540540541 * texture2D(inputTex, vec2(tc.x, tc.y + 3 * inputTexelSize.y));
		sum += 0.0162162162 * texture2D(inputTex, vec2(tc.x, tc.y + 4 * inputTexelSize.y));
		
		fragData = sum;

	}
	`
/*
08. 
09.float3 Uncharted2Tonemap(float3 x)
10.{
11.return ((x*(A*x+C*B)+D*E)/(x*(A*x+B)+D*F))-E/F;
12.}
13. 
14.float4 ps_main( float2 texCoord  : TEXCOORD0 ) : COLOR
15.{
16.float3 texColor = tex2D(Texture0, texCoord );
17.texColor *= 16;  // Hardcoded Exposure Adjustment
18. 
19.float ExposureBias = 2.0f;
20.float3 curr = Uncharted2Tonemap(ExposureBias*texColor);
21. 
22.float3 whiteScale = 1.0f/Uncharted2Tonemap(W);
23.float3 color = curr*whiteScale;
24. 
25.float3 retColor = pow(color,1/2.2);
26.return float4(retColor,1);
27.}
*/
const blendFragShaderSrc = `
	#version 430
	noperspective in vec2 vTc;

	layout (location = 0) out vec4 fragData;

	layout (location = 0) uniform sampler2D sceneTex;
	layout (location = 1) uniform sampler2D blurredTex;

	vec3 toneMap(vec2 tc) {
		vec3 texColor = texture2D(sceneTex, tc).rgb;
		texColor *= 0.5; // exposure

		vec3 x = max(vec3(0), texColor - 0.004);
		vec3 toneMapped = (x*(6.2*x+.5))/(x*(6.2*x+1.7)+0.06);
		
		return toneMapped;
	}

	void main(void) {
		vec3 blurredColor = texture2D(blurredTex, vTc).rgb;
		fragData = vec4(toneMap(vTc) + blurredColor, 1);

		if (vTc.x > 0.4 && vTc.y > 0.8) {
			float y = 5 * (vTc.y - 0.8);

			if (vTc.x > 0.8) {
				fragData = texture2D(blurredTex, vec2(5*(vTc.x-0.8), y));
			} else if (vTc.x > 0.6) {
				fragData = vec4(toneMap(vec2(5 * (vTc.x - 0.6), y)), 1);
			} else {
				fragData = texture2D(sceneTex, vec2(5 * (vTc.x - 0.4), y));
			}
		}
	}
`


type BlurFilter struct {
	PostProcessFilterBase

	blurFBOs [2]gl.Uint
	blurTexs [2]gl.Uint

	downSampleShader *shader.Shader
	blurXShader *shader.Shader
	blurYShader *shader.Shader
	blendShader *shader.Shader


}




func MakeBlurFilter(w, h int) (b *BlurFilter) {
	b = new(BlurFilter)
	b.PostProcessFilterBase.init(w,h)


	// blur FBOs
	for i := 0; i < 2; i++ {
		gl.GenTextures(1, &(b.blurTexs[i]))

		gl.BindTexture(gl.TEXTURE_2D, b.blurTexs[i]);
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, gl.Sizei(w) / 4, gl.Sizei(h) / 4, 0, gl.RGBA, gl.FLOAT, nil)
		setTexParameters()

	
		gl.BindTexture(gl.TEXTURE_2D, 0)
	
		gl.GenFramebuffers(1, &(b.blurFBOs[i]))
		gl.BindFramebuffer(gl.FRAMEBUFFER, b.blurFBOs[i])
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, b.blurTexs[i], 0)
	
		drawBufs := []gl.Enum{ gl.COLOR_ATTACHMENT0 }
		gl.DrawBuffers(1, &(drawBufs[0]))

		checkFramebuffer()

		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	}

	b.downSampleShader = shader.Make()
	b.downSampleShader.AddShaderSource(vtxShaderSrc, gl.VERTEX_SHADER)
	b.downSampleShader.AddShaderSource(downSampleFragShaderSrc, gl.FRAGMENT_SHADER)
	b.downSampleShader.Link()

	b.blurXShader = shader.Make()
	b.blurXShader.AddShaderSource(vtxShaderSrc, gl.VERTEX_SHADER)
	b.blurXShader.AddShaderSource(blurXFragShaderSrc, gl.FRAGMENT_SHADER)
	b.blurXShader.Link()


	b.blurYShader = shader.Make()
	b.blurYShader.AddShaderSource(vtxShaderSrc, gl.VERTEX_SHADER)
	b.blurYShader.AddShaderSource(blurYFragShaderSrc, gl.FRAGMENT_SHADER)
	b.blurYShader.Link()
	

	b.blendShader = shader.Make()
	b.blendShader.AddShaderSource(vtxShaderSrc, gl.VERTEX_SHADER)
	b.blendShader.AddShaderSource(blendFragShaderSrc, gl.FRAGMENT_SHADER)
	b.blendShader.Link()
	
	
	return
}

func (b *BlurFilter) Delete() {
	b.PostProcessFilterBase.delete()
}
/*
func (b *BlurFilter) BeginRender() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, b.inputFBO)
	gl.Viewport(0,0, gl.Sizei(b.w), gl.Sizei(b.h))
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

func (b *BlurFilter) EndRender() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}
*/

func (b *BlurFilter) Apply(gbuf *gbuffer.GBuffer, inputTex gl.Uint, P, V *vmath.Matrix4) (outputTex gl.Uint) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, b.blurFBOs[0])
	
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, inputTex);

	b.downSampleShader.ProgramUniform1i(0, 0)
	b.downSampleShader.ProgramUniform2f(1, 1.0 / float32(b.w), 1.0 / float32(b.h))

	b.downSampleShader.Enable()
	b.fsQuadVAO.Draw() // downsample input texture into blurFBOs[0]
	b.downSampleShader.Disable()


	
	for i := 0; i < 4; i++ {
	gl.BindFramebuffer(gl.FRAMEBUFFER, b.blurFBOs[1])

	gl.BindTexture(gl.TEXTURE_2D, b.blurTexs[0])
	b.blurXShader.ProgramUniform1i(0, 0)
	b.blurXShader.ProgramUniform2f(1, 1.0 / float32(b.w/4), 1.0 / float32(b.h/4))

	b.blurXShader.Enable()
	b.fsQuadVAO.Draw()
	b.blurXShader.Disable()

	gl.BindFramebuffer(gl.FRAMEBUFFER, b.blurFBOs[0])

	gl.BindTexture(gl.TEXTURE_2D, b.blurTexs[1])
	b.blurYShader.ProgramUniform1i(0, 0)
	b.blurYShader.ProgramUniform2f(1, 1.0 / float32(b.w/4), 1.0 / float32(b.h/4))

	b.blurYShader.Enable()
	b.fsQuadVAO.Draw()
	b.blurYShader.Disable()
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, b.outputFBO)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, inputTex)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, b.blurTexs[0])

	b.blendShader.ProgramUniform1i(0,0)
	b.blendShader.ProgramUniform1i(1,1)

	b.blendShader.Enable()
	b.fsQuadVAO.Draw()
	b.blendShader.Disable()

	gl.BindTexture(gl.TEXTURE_2D, 0);
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, 0);

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	return b.outputTex

/*
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0) // keep reading from b.blurFBOs[1] but write to default FBO
	gl.BlitFramebuffer(0, 0, gl.Int(b.w/4), gl.Int(b.h/4), 0, 0, gl.Int(b.w), gl.Int(b.h), gl.COLOR_BUFFER_BIT, gl.LINEAR)
*/

}



	
