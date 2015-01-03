package shader

import (
	"C"
	"unsafe"
	"fmt"
	"github.com/go-gl/glow/gl-core/4.1/gl"
	vmath "github.com/rwesterteiger/vectormath"
	"strings"
	"reflect"
	"log"
)

type Shader struct {
	enabled bool
	program uint32 // just the program handle for now
}

func Make() *Shader {
	s := &Shader{program: gl.CreateProgram()}

	return s
}

func (s *Shader) AddShaderSource(src string, sType uint32) {
	srcString := gl.Str(src + "\x00")

	obj := gl.CreateShader(sType)
	defer gl.DeleteShader(obj)

	gl.ShaderSource(obj, 1, &srcString, nil)
	gl.CompileShader(obj)

	var success int32
	gl.GetShaderiv(obj, gl.COMPILE_STATUS, &success)

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

	var success int32
	gl.GetProgramiv(s.program, gl.LINK_STATUS, &success)

	if success == 0 {
		fmt.Printf("Error linking shader!")
		s.printShaderLog(s.program)
		panic("Shader linking error")
	}
}

func (s *Shader) BindFragDataLocation(idx uint32, name string) {
	glName := gl.Str(name + "\x00")

	gl.BindFragDataLocation(s.program, idx, glName)
}

func (s *Shader) BindAttribLocation(idx uint32, name string) {
	glName := gl.Str(name + "\x00")

	gl.BindAttribLocation(s.program, idx, glName)
}

func (s *Shader) Enable() {
	gl.UseProgram(s.program)
	s.enabled = true
}

func (s *Shader) Disable() {
	gl.UseProgram(0)
	s.enabled = false
}

func (s *Shader) withEnabled(f func()) {
	if (!s.enabled) {
		s.Enable()
		defer s.Disable()
	}

	f()
}

func (s *Shader) setM4Uniform(location int32, m *vmath.Matrix4) {
	s.withEnabled(func() {
		floatData := make([]float32, 16)

		for row := 0; row < 4; row++ {
			for col := 0; col < 4; col++ {
				floatData[4*col+row] = float32(m.GetElem(col, row))
			}
		}

		gl.UniformMatrix4fv(int32(location), 1, false, &floatData[0])
	})
}

func (s *Shader) ProgramUniformM4(location int, m *vmath.Matrix4) {
	s.withEnabled(func() { s.setM4Uniform(int32(location), m) })
}

func (s *Shader) ProgramUniformF4(location int, v *vmath.Vector4) {
	s.withEnabled(func() { gl.Uniform4f(int32(location), float32(v.X), float32(v.Y), float32(v.Z), float32(v.W)) })
}

func (s *Shader) ProgramUniform1f(location int, x float32) {
	s.withEnabled(func() { gl.Uniform1f(int32(location), float32(x)) })
}

func (s *Shader) ProgramUniform2f(location int, x float32, y float32) {
	gl.Uniform2f(int32(location), float32(x), float32(y))
}

func (s *Shader) ProgramUniform3f(location int, x, y, z float32) {
	s.withEnabled(func() { gl.Uniform3f(int32(location), float32(x), float32(y), float32(z)) })
}

func (s *Shader) ProgramUniform4f(location int, x, y, z, w float32) {
	s.withEnabled(func() { gl.Uniform4f(int32(location), float32(x), float32(y), float32(z), float32(w)) })
}

func (s *Shader) ProgramUniform1i(location int, x int) {
	s.withEnabled(func() { gl.Uniform1i(int32(location), int32(x)) })
}

// uStruct is assumed to be a pointer to a struct having fields containing uniform values tagged like this:
//
// type ambientLightUniforms struct {
//	albedoTex int		`glUniform : "albedoTex"`
// }

func (s *Shader) SetUniforms(uStruct interface{}) {
	s.withEnabled(func() {

		v := reflect.ValueOf(uStruct).Elem() // deref interface, then pointer
		// t := reflect.TypeOf(v)

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			fieldType  := v.Type().Field(i)
			tag := fieldType.Tag
		
			uniformName := tag.Get("glUniform")

			if uniformName == "" {
				uniformName = fieldType.Name // if there is no `glUniform:"<name>"` tag, use the field name as uniform name
			}

			loc := int32(s.GetUniformLocation(uniformName))

			if (loc == -1) {
				log.Fatal(fmt.Sprintf("Shader.SetUniforms(): Unknown uniform \"%v\"!", uniformName))
			}
		
			switch x := fieldValue.Interface().(type) {
			default:
				panic(fmt.Sprintf("SetUniform: Unexpected type %T!\n", x))
			case int:
				gl.Uniform1i(loc, int32(x))
			case float32:
				gl.Uniform1f(loc, float32(x))
			case vmath.Vector3:
				gl.Uniform3f(loc, float32(x.X), float32(x.Y), float32(x.Z))
			case vmath.Vector4:
				gl.Uniform4f(loc, float32(x.X), float32(x.Y), float32(x.Z), float32(x.W))
			case vmath.Matrix4:
				s.setM4Uniform(loc, &x)
			}
		}
	})
}

func (s *Shader) printShaderLog(obj uint32) {
	var logLength int32
	logLength = 1
	gl.GetShaderiv(obj, gl.INFO_LOG_LENGTH, &logLength)

	log := make([]byte, logLength+1) // make sure it zero terminated

	var result int32

	gl.GetShaderInfoLog(obj, int32(logLength), &result, (*uint8)(unsafe.Pointer(&log[0])))

	if result == 0 {
		fmt.Println("Unable to retrieve shader info log!")
		return
	}


	fmt.Printf("Shader error log:\n%v\n", string(log))
}

func (s *Shader) GetUniformLocation(name string) int {
	glName := gl.Str(name + "\x00")
	return int(gl.GetUniformLocation(s.program, glName))
}

