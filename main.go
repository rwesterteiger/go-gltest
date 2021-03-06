package main

import (
	gl "github.com/chsc/gogl/gl43"
	"github.com/jteeuwen/glfw"
	//	"github.com/rwesterteiger/vectormath"
	"log"
	"github.com/rwesterteiger/go-gltest/shader"
	"github.com/rwesterteiger/go-gltest/buffers"
	"github.com/rwesterteiger/go-gltest/geom"
	vmath "github.com/rwesterteiger/vectormath"
	"math"
	"fmt"

	"github.com/rwesterteiger/go-gltest/scene"
	"github.com/rwesterteiger/go-gltest/lights"
	"github.com/rwesterteiger/go-gltest/post"
	"time"	
)

const (
	Title  = "Hello Shader"
	Width  = 1024
	Height = 768
)



const vertexShaderSource = `
#version 430
layout (location = 0) in vec3 vtxPos;
layout (location = 1) in vec2 texCoord;
out vec3 vPosition;
out vec2 vTexCoord;

void main(void) {
	vPosition = vtxPos;
	vTexCoord = texCoord;
}
`


const fragmentShaderSource = `
#version 430
layout (location = 0) out vec4 fragData;
in vec3 tcNormal;

void main(void)
{
	float diffuse = max(0.0, dot(vec3(1,1,0), tcNormal));
	fragData = diffuse * vec4(1.0, 0.0, 0.0, 1.0);
}
`


func makePlaneVAO() (*buffers.VAO) {
	vtxs := buffers.MakeVBOFromVec3s([]vmath.Vector3{ {-10, 0, 10}, {10, 0, 10}, {10, 0, -10}, {-10, 0, -10 } })
	normals := buffers.MakeVBOFromVec3s([]vmath.Vector3{ {0,1,0}, {0,1,0}, {0,1,0}, {0,1,0} })
	indices := []uint32{ 0, 1, 2, 2, 3, 0 }

	vao := buffers.MakeVAO(gl.TRIANGLES, 6)
	vao.AttachVBO(0, vtxs)
	vao.AttachVBO(1, normals)
	vao.SetIndexBuffer(indices)

	return vao
}
	
func makeQuadShader() (quadShader *shader.Shader) {
	const quadVtxShaderSrc =`
	#version 430
	layout (location = 0) in vec2 vtx;
	layout (location = 1) in vec2 tc;

	out vec2 vTc;

	void main(void) {
		gl_Position = vec4(vtx.xy, 0, 1);
		vTc = tc;
	}
	`

	const quadFragShaderSrc = `
	#version 430
	layout (location = 0) out vec4 fragData;
	in vec2 vTc;

	layout (location = 0) uniform sampler2D albedoTex;
	layout (location = 1) uniform sampler2D normalTex;
	layout (location = 2) uniform sampler2D depthTex;

	void main(void)
	{
		float foo = mod(vTc.x, 1.0 / 3.0);

		vec2 tc = vec2(3 * mod(vTc.x, 1.0 / 3.0), vTc.y);

		//float z = texture2D(depthTex, vTc).x;
		//fragData = texture2D(albedoTex, vTc);
		//vec3 n = texture2D(normalTex, vTc).xyz;

		switch (int(vTc.x * 3)) {
			case 0:
				fragData = texture2D(albedoTex, tc);
				break;
			case 1:
				fragData = vec4(pow(texture2D(depthTex, tc).x, 16));
				break;
			case 2:
				fragData = texture2D(normalTex, tc) /2 + 0.5;
				break;
		}

	}
	`

	quadShader = shader.Make()

	quadShader.AddShaderSource(quadVtxShaderSrc, gl.VERTEX_SHADER)
	quadShader.AddShaderSource(quadFragShaderSrc, gl.FRAGMENT_SHADER)
	quadShader.Link()

	return
}

func makeAmbientBlitShader() (*shader.Shader) {
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

	layout (location = 0) uniform sampler2D albedoTex;

	void main(void)
	{
		fragData = 0.1 * texture2D(albedoTex, vTc);
	}
	`

	s := shader.Make()

	s.AddShaderSource(vSrc, gl.VERTEX_SHADER)
	s.AddShaderSource(fSrc, gl.FRAGMENT_SHADER)
	s.Link()

	return s
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

func makeQuadVAO() (*buffers.VAO) {
	vtxs := buffers.MakeVBOFromVec2s([]vmath.Vector2{ { -0.2, -0.98}, {0.98, -0.98}, {0.98, -0.6}, {-0.2, -0.6} })
	tcs := buffers.MakeVBOFromVec2s([]vmath.Vector2{ { 0,0 }, {1,0}, {1,1}, {0,1} })
	indices := []uint32{ 0, 1, 2, 2, 3, 0 }

	vao := buffers.MakeVAO(gl.TRIANGLES, 6)
	vao.AttachVBO(0, vtxs)
	vao.AttachVBO(1, tcs)
	vao.SetIndexBuffer(indices)

	return vao
}
/*
func makeSceneRenderer() func(float32,  *vmath.Matrix4, *vmath.Matrix4) {
	objVAO := geom.LoadOBJ("monkey.obj")
	planeVAO := makePlaneVAO()

	var blenderToGLXForm vmath.Transform3
	vmath.T3MakeFromCols(&blenderToGLXForm, &vmath.Vector3{1,0,0}, &vmath.Vector3{0,1,0}, &vmath.Vector3{0,0,1}, &vmath.Vector3{0,0,0})

	return func(t float32, camProjectionMatrix *vmath.Matrix4, camViewMatrix *vmath.Matrix4) {	

		var objSpinRotation vmath.Transform3
		vmath.T3MakeRotationY(&objSpinRotation, 0.0) // spin around

		var objTranslation vmath.Transform3
		vmath.T3MakeTranslation(&objTranslation, &vmath.Vector3{0, 0.25, 0})

		var modelMatrixObj vmath.Matrix4
		vmath.M4MakeFromT3(&modelMatrixObj, &objSpinRotation)
		vmath.M4MulT3(&modelMatrixObj, &modelMatrixObj, &objTranslation)

		vmath.M4MulT3(&modelMatrixObj, &modelMatrixObj, &blenderToGLXForm)

		var modelMatrixPlane vmath.Matrix4
		vmath.M4MakeIdentity(&modelMatrixPlane)

		sh.ProgramUniformM4(0, camProjectionMatrix)
		sh.ProgramUniformM4(4, camViewMatrix)

		sh.BindFragDataLocation(0, "fragAlbedo")
		sh.BindFragDataLocation(1, "fragNormal")

		sh.Enable()

		sh.ProgramUniformM4(8, &modelMatrixPlane)
		sh.ProgramUniformF4(12, vmath.Vector4{1.0, 1.0, 1.0, 1})

		planeVAO.Draw()

		sh.ProgramUniformM4(8, &modelMatrixObj)
		sh.ProgramUniformF4(12, vmath.Vector4{1.0, 1.0, 1.0, 1.0})
		objVAO.Draw()

		sh.Disable()
	}
}

*/
func main() {
	if err := glfw.Init(); err != nil {
		log.Fatal(err)
	}
	defer glfw.Terminate()

	glfw.OpenWindowHint(glfw.OpenGLVersionMajor, 4)
	glfw.OpenWindowHint(glfw.OpenGLVersionMinor, 3)
	glfw.OpenWindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile);
	glfw.OpenWindowHint(glfw.WindowNoResize, 1)
	glfw.SetSwapInterval(0)

	if err := glfw.OpenWindow(Width, Height, 0, 0, 0, 0, 32, 0, glfw.Windowed); err != nil {
		log.Fatal(err)
	}
	defer glfw.CloseWindow()


	glfw.SetWindowTitle(Title)

	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}
	
	//quadShader := makeQuadShader()
	//quadVAO := makeQuadVAO()


	//ambientBlitShader := makeAmbientBlitShader()
	//fsQuadVAO := makeFullscreenQuadVAO()

	scene := scene.Make(Width, Height)
	defer scene.Delete()

	var zNear float32 = 0.1
	var zFar float32 = 100.0
	scene.SetCameraPerspective(60.0 / 360.0 * 2 * math.Pi, float32(Width) / float32(Height), zNear, zFar)



	var blenderToGLXForm vmath.Transform3
	vmath.T3MakeFromCols(&blenderToGLXForm, &vmath.Vector3{1,0,0}, &vmath.Vector3{0,1,0}, &vmath.Vector3{0,0,1}, &vmath.Vector3{0,0,0})

	var objTranslation vmath.Transform3
	vmath.T3MakeTranslation(&objTranslation, &vmath.Vector3{0, 0.25, 0})

	var modelMatrixMonkey vmath.Matrix4
	vmath.M4MakeFromT3(&modelMatrixMonkey, &objTranslation)
	vmath.M4MulT3(&modelMatrixMonkey, &modelMatrixMonkey, &blenderToGLXForm)


	monkey := geom.LoadOBJ("monkey.obj", &vmath.Vector4{1,1,1,1})
	monkey.SetModelMatrix(&modelMatrixMonkey)
	scene.AddObject(monkey)


	var monkeyArrayTrans vmath.Transform3
	vmath.T3MakeTranslation(&monkeyArrayTrans, &vmath.Vector3{0,0,-2})
	vmath.M4MulT3(&modelMatrixMonkey, &modelMatrixMonkey, &monkeyArrayTrans)

	monkey = geom.LoadOBJ("monkey.obj", &vmath.Vector4{1,1,1,1})
	monkey.SetModelMatrix(&modelMatrixMonkey)
	scene.AddObject(monkey)

	vmath.T3MakeTranslation(&monkeyArrayTrans, &vmath.Vector3{0,0,2})
	vmath.M4MulT3(&modelMatrixMonkey, &modelMatrixMonkey, &monkeyArrayTrans)

	vmath.M4MulT3(&modelMatrixMonkey, &modelMatrixMonkey, &monkeyArrayTrans)
	monkey = geom.LoadOBJ("monkey.obj", &vmath.Vector4{1,1,1,1})
	monkey.SetModelMatrix(&modelMatrixMonkey)
	scene.AddObject(monkey)

	scene.AddObject(geom.MakeObject(makePlaneVAO(), &vmath.Vector4{1,1,1,1}))

	scene.AddLight(lights.MakeAmbientLight())

	scene.AddLight(lights.MakeSpotLight(&vmath.Point3{0, 3,-2}, &vmath.Point3{0,0,-2}, &vmath.Vector3{0,0,-1}, 2, &vmath.Vector3{0.5,0,0}))
	scene.AddLight(lights.MakeSpotLight(&vmath.Point3{0, 3, 0}, &vmath.Point3{0,0, 0}, &vmath.Vector3{0,0,-1}, 2, &vmath.Vector3{0,0.5,0}))
	scene.AddLight(lights.MakeSpotLight(&vmath.Point3{0, 3, 2}, &vmath.Point3{0,0, 2}, &vmath.Vector3{0,0,-1}, 2, &vmath.Vector3{0,0,0.5}))

	//scene.AddLight(lights.MakeSpotLight(&vmath.Point3{2, 2, 2}, &vmath.Point3{0,0.0,0}, &vmath.Vector3{0,0,-1}, 1.5, &vmath.Vector3{0.5,0,0}))
	//scene.AddLight(lights.MakeSpotLight(&vmath.Point3{-2,2, 2}, &vmath.Point3{0,0.0,0}, &vmath.Vector3{0,0,-1}, 1.5, &vmath.Vector3{0,0.5,0}))
	//scene.AddLight(lights.MakeSpotLight(&vmath.Point3{0, 2,-2}, &vmath.Point3{0,0.0,0}, &vmath.Vector3{0,0,-1}, 1.5, &vmath.Vector3{0,0,0.5}))

	dofFilter := post.MakeDoFFilter(Width, Height, 3.4)
	scene.AddPostFilter(dofFilter)

	blurFilter := post.MakeBlurFilter(Width, Height)
	scene.AddPostFilter(blurFilter)

	var t float32 = 0.0

	startTime := time.Now()
	frameCount := 0
	for glfw.WindowParam(glfw.Opened) == 1 {
		if frameCount % 1000 == 0 {
			thisFrameTime := time.Now()
			seconds := thisFrameTime.Sub(startTime).Seconds() / 1000.0
			fmt.Printf("Frametime: %4.1f ms (%4.1f fps)\n", 1000.0 * seconds, 1.0 / seconds)

			startTime = thisFrameTime
		}

		camX := float32(-3 * math.Sin(float64(t)))
		camZ := 0.7 +  float32(-3 * math.Cos(float64(t)))

		scene.SetCameraLookAt(&vmath.Point3{camX, 2, camZ}, &vmath.Point3{0,0.6,0.7}, &vmath.Vector3{0,1,0})

		gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
		scene.Render()

		//blurFilter.BeginRender()

		//gl.ActiveTexture(gl.TEXTURE0)
		//gl.BindTexture(gl.TEXTURE_2D, gbuf.GetAlbedoTex())

		//ambientBlitShader.ProgramUniform1i(0,0)
		//ambientBlitShader.Enable()


		//fsQuadVAO.Draw()

		//ambientBlitShader.Disable()


		//blurFilter.EndRender()
		//blurFilter.DisplayBlurResult()

		// bind textures
/*
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, gbuf.GetAlbedoTex())

		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, gbuf.GetNormalTex())

		gl.ActiveTexture(gl.TEXTURE2)
		gl.BindTexture(gl.TEXTURE_2D, gbuf.GetDepthTex())
	
		quadShader.ProgramUniform1i(0, 0)
		quadShader.ProgramUniform1i(1, 1)
		quadShader.ProgramUniform1i(2, 2)

		quadShader.Enable()
		quadVAO.Draw()
		quadShader.Disable()

		gl.BindTexture(gl.TEXTURE_2D, 0)
		gl.ActiveTexture(gl.TEXTURE2)
		gl.BindTexture(gl.TEXTURE_2D, 0)
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, 0)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, 0)

		
	
		
*/

	t = 16.4;
		
		//t = t + 0.03 / 360.0 * 2 * math.Pi 
		glfw.SwapBuffers()
		frameCount++
	}
}

