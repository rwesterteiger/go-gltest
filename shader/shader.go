package shader

import (
	"C"
	"unsafe"
	"fmt"
	gl "github.com/rwesterteiger/gogl/gl32"
	vmath "github.com/rwesterteiger/vectormath"
	"strings"
	"reflect"
	// "log"
)

type Shader struct {
	enabled bool
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

func (s *Shader) setM4Uniform(location gl.Int, m *vmath.Matrix4) {
	s.withEnabled(func() {
		floatData := make([]gl.Float, 16)

		for row := 0; row < 4; row++ {
			for col := 0; col < 4; col++ {
				floatData[4*col+row] = gl.Float(m.GetElem(col, row))
			}
		}

		gl.UniformMatrix4fv(gl.Int(location), 1, gl.FALSE, &floatData[0])
	})
}

func (s *Shader) ProgramUniformM4(location int, m *vmath.Matrix4) {
	s.withEnabled(func() { s.setM4Uniform(gl.Int(location), m) })
}

func (s *Shader) ProgramUniformF4(location int, v *vmath.Vector4) {
	s.withEnabled(func() { gl.Uniform4f(gl.Int(location), gl.Float(v.X), gl.Float(v.Y), gl.Float(v.Z), gl.Float(v.W)) })
}

func (s *Shader) ProgramUniform1f(location int, x float32) {
	s.withEnabled(func() { gl.Uniform1f(gl.Int(location), gl.Float(x)) })
}

func (s *Shader) ProgramUniform2f(location int, x float32, y float32) {
	gl.Uniform2f(gl.Int(location), gl.Float(x), gl.Float(y))
}

func (s *Shader) ProgramUniform3f(location int, x, y, z float32) {
	s.withEnabled(func() { gl.Uniform3f(gl.Int(location), gl.Float(x), gl.Float(y), gl.Float(z)) })
}

func (s *Shader) ProgramUniform4f(location int, x, y, z, w float32) {
	s.withEnabled(func() { gl.Uniform4f(gl.Int(location), gl.Float(x), gl.Float(y), gl.Float(z), gl.Float(w)) })
}

func (s *Shader) ProgramUniform1i(location int, x int) {
	s.withEnabled(func() { gl.Uniform1i(gl.Int(location), gl.Int(x)) })
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
			// f := v.Field(i)

			//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", fieldType.Name, fieldValue.Interface(), tag.Get("glUniform"))

			uniformName := tag.Get("glUniform")

			if uniformName == "" { // non-tagged field, ignore
				continue
			}

			loc := gl.Int(s.GetUniformLocation(uniformName))
			//fmt.Println(fieldValue.Interface())

			switch x := fieldValue.Interface().(type) {
			default:
				panic(fmt.Sprintf("SetUniform: Unexpected type %T!\n", x))
			case int:
				gl.Uniform1i(loc, gl.Int(x))
			case float32:
				gl.Uniform1f(loc, gl.Float(x))
			case vmath.Vector4:
				gl.Uniform4f(loc, gl.Float(x.X), gl.Float(x.Y), gl.Float(x.Z), gl.Float(x.W))
			case vmath.Matrix4:
				s.setM4Uniform(loc, &x)
			}
		}
	})
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

