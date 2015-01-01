package shader

import (
	"C"
	"unsafe"
	"fmt"
	gl "github.com/rwesterteiger/gogl/gl32"
	vmath "github.com/rwesterteiger/vectormath"
	"strings"
	// "log"
)

type Shader struct {
	program gl.Uint // just the program handle for now
}

func Make() *Shader {
	s := &Shader{program: gl.CreateProgram()}

	return s
}

func (s *Shader) AddShaderSource(src string, sType gl.Enum) {
	srcString := gl.GLString(src)
	defer gl.GLStringFree(srcString)

	obj := gl.CreateShader(sType)
	defer gl.DeleteShader(obj)

	gl.ShaderSource(obj, 1, &srcString, nil)
	gl.CompileShader(obj)

	var success gl.Int
	gl.GetShaderiv(obj, gl.COMPILE_STATUS, &success)

	fmt.Printf("compile result = %v\n", success)

	if success == 0 {
		fmt.Println("Error compiling shader:")

		lines := strings.Split(src, "\n")
	
		for i,s := range(lines) {
			fmt.Printf("%10d %v\n", i+1, s)
		}
		fmt.Println("")

		s.printShaderLog(obj)
		panic("Shader compilation error")
	}
	gl.AttachShader(s.program, obj)
}

func (s *Shader) Delete() {
	gl.DeleteProgram(s.program)
}

func (s *Shader) Link() {
	gl.LinkProgram(s.program)

	var success gl.Int
	gl.GetProgramiv(s.program, gl.LINK_STATUS, &success)

	fmt.Printf("link result = %v\n", success)

	if success == 0 {
		fmt.Printf("Error linking shader!")
		s.printShaderLog(s.program)
		panic("Shader linking error")
	}
}

func (s *Shader) BindFragDataLocation(idx gl.Uint, name string) {
	glName := gl.GLString(name)
	defer gl.GLStringFree(glName)

	gl.BindFragDataLocation(s.program, idx, glName)
}

func (s *Shader) BindAttribLocation(idx gl.Uint, name string) {
	glName := gl.GLString(name)
	defer gl.GLStringFree(glName)

	gl.BindAttribLocation(s.program, idx, glName)
}

func (s *Shader) Enable() {
	gl.UseProgram(s.program)
}

func (_ *Shader) Disable() {
	gl.UseProgram(0)
}

func (s *Shader) ProgramUniformM4(location int, m *vmath.Matrix4) {
	floatData := make([]gl.Float, 16)

	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			floatData[4*col+row] = gl.Float(m.GetElem(col, row))
		}
	}

	gl.UniformMatrix4fv(gl.Int(location), 1, gl.FALSE, &floatData[0])
}

func (s *Shader) ProgramUniformF4(location int, v *vmath.Vector4) {
	s.Enable()
	gl.Uniform4f(gl.Int(location), gl.Float(v.X), gl.Float(v.Y), gl.Float(v.Z), gl.Float(v.W))
	s.Disable()
}

func (s *Shader) ProgramUniform1f(location int, x float32) {
	s.Enable()
	gl.Uniform1f(gl.Int(location), gl.Float(x))
	s.Disable()
}

func (s *Shader) ProgramUniform2f(location int, x float32, y float32) {
	s.Enable()
	gl.Uniform2f(gl.Int(location), gl.Float(x), gl.Float(y))
	s.Disable()
}

func (s *Shader) ProgramUniform3f(location int, x, y, z float32) {
	s.Enable()
	gl.Uniform3f(gl.Int(location), gl.Float(x), gl.Float(y), gl.Float(z))
	s.Disable()
}

func (s *Shader) ProgramUniform4f(location int, x, y, z, w float32) {
	s.Enable()
	gl.Uniform4f(gl.Int(location), gl.Float(x), gl.Float(y), gl.Float(z), gl.Float(w))
	s.Disable()
}

func (s *Shader) ProgramUniform1i(location int, x int) {
	s.Enable()
	gl.Uniform1i(gl.Int(location), gl.Int(x))
	s.Disable()
}

func (s *Shader) printShaderLog(obj gl.Uint) {
	var logLength gl.Int
	logLength = 1
	gl.GetShaderiv(obj, gl.INFO_LOG_LENGTH, &logLength)

	log := make([]byte, logLength+1) // make sure it zero terminated

	var result gl.Sizei

	gl.GetShaderInfoLog(obj, gl.Sizei(logLength), &result, (*gl.Char)(unsafe.Pointer(&log[0])))

	if result == 0 {
		fmt.Println("Unable to retrieve shader info log!")
		return
	}


	fmt.Printf("Shader error log:\n%v\n", string(log))
}

func (s *Shader) GetUniformLocation(name string) int {
	glName := gl.GLString(name)
	defer gl.GLStringFree(glName)

	return int(gl.GetUniformLocation(s.program, glName))
}

