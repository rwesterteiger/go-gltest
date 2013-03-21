package shader
import (
	gl "github.com/chsc/gogl/gl43"
	vmath "github.com/rwesterteiger/vectormath"
	"fmt"
	"log"
)

type Shader struct {
	program gl.Uint // just the program handle for now
}

func Make() (*Shader) {
	s := &Shader{ program : gl.CreateProgram() }

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
		log.Fatal("Shader compilation failed!");
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
		log.Fatal("Shader linking failed!");
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
			floatData[4 * col + row] = gl.Float(m.GetElem(col, row))
		}
	}

	gl.ProgramUniformMatrix4fv(s.program, gl.Int(location), 1, gl.FALSE, &floatData[0])
}


func (s *Shader) ProgramUniformF4(location int, v *vmath.Vector4) {
	gl.ProgramUniform4f(s.program, gl.Int(location), gl.Float(v.X), gl.Float(v.Y), gl.Float(v.Z), gl.Float(v.W))
}

func (s *Shader) ProgramUniform1f(location int, x float32) {
	gl.ProgramUniform1f(s.program, gl.Int(location), gl.Float(x))
}

func (s *Shader) ProgramUniform2f(location int, x float32, y float32) {
	gl.ProgramUniform2f(s.program, gl.Int(location), gl.Float(x), gl.Float(y))
}

func (s *Shader) ProgramUniform3f(location int, x,y,z float32) {
	gl.ProgramUniform3f(s.program, gl.Int(location), gl.Float(x), gl.Float(y), gl.Float(z))
}


func (s *Shader) ProgramUniform4f(location int, x,y,z,w float32) {
	gl.ProgramUniform4f(s.program, gl.Int(location), gl.Float(x), gl.Float(y), gl.Float(z), gl.Float(w))
}

func (s *Shader) ProgramUniform1i(location int, x int) {
	gl.ProgramUniform1i(s.program, gl.Int(location), gl.Int(x))
}

