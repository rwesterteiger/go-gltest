package geom
import (
	"github.com/go-gl/glow/gl-core/4.1/gl"
	"github.com/rwesterteiger/go-gltest/buffers"
	"github.com/rwesterteiger/go-gltest/util"

	"log"
	"strings"
	"strconv"
	"os"
	"bufio"
	"io"
	"fmt"
	vmath "github.com/rwesterteiger/vectormath"
)

func parseFloat(s string) float32 {
	result, err := strconv.ParseFloat(s, 32)

	if (err != nil) {
		log.Fatalf("Error parsing floating point number: %s\n", s)
	}

	return float32(result)
}

type OBJFace struct {
	VtxIndices [3]uint32
	NormalIndices [3]uint32
}

type OBJData struct {
	Vertices []vmath.Vector3
	Normals []vmath.Vector3
	Faces []OBJFace
}

func parseOBJFace(args []string) (f OBJFace) {
	for i, a := range(args) {
		fields := strings.Split(a, "/")

		n, err := strconv.Atoi(fields[0])
		if (err != nil) {
			log.Fatal(err)
		}
		f.VtxIndices[i] = uint32(n)

		n, err = strconv.Atoi(fields[2])
		if (err != nil) {
			log.Fatal(err)
		}

		f.NormalIndices[i] = uint32(n)
	}

	return
}

func parseOBJ(path string) (objData *OBJData) {
	objData = new(OBJData)

	f, err := os.Open(path)
	defer f.Close()

	if (err != nil) {
		log.Fatal(err)
	}

	fileBuf := bufio.NewReader(f)

	for {
		line, err := fileBuf.ReadString('\n')

		if err == nil {
			line = line[:len(line)-1] // remove trailing newline
		} else {
			if err != io.EOF {
				log.Fatal(err)
			}
		}
		fields := strings.Fields(line)
		//fmt.Printf("fields (%d): %v\n", len(fields), fields)

		if len(fields) > 0 {
			switch fields[0] {
				case "#":
				case "v":
					vec := vmath.Vector3{ parseFloat(fields[1]), parseFloat(fields[2]), parseFloat(fields[3]) }
					objData.Vertices = append(objData.Vertices, vec)
				case "vn":
					vec := vmath.Vector3{ parseFloat(fields[1]), parseFloat(fields[2]), parseFloat(fields[3]) }
					objData.Normals = append(objData.Normals, vec)
				case "f":
					objData.Faces = append(objData.Faces, parseOBJFace(fields[1:]))
				case "o":
					fmt.Printf("object name: %s\n", fields[1])
				case "s":
				default:
					log.Fatalf("Unable to parse line: %s\n", line)
			}
				
		}

		if err == io.EOF {
			break
		}
	}		

	return
}

func makeVAOFromOBJ(obj *OBJData) (*buffers.VAO) {
	indicesToIdxMap := make(map[string]uint32) // maps a string "vtxidx/normalidx" to the index in our vertex and normal vbos

	idxBuffer := make([]uint32, 0)
	vtxBuffer := make([]vmath.Vector3, 0)
	normalBuffer := make([]vmath.Vector3, 0)

	for _, f := range(obj.Faces) {
		for i := 0; i < 3; i++ {
			key := fmt.Sprintf("%d/%d", f.VtxIndices[i], f.NormalIndices[i])
			idx, ok := indicesToIdxMap[key]

			if !ok {
				idx = uint32(len(vtxBuffer))
				indicesToIdxMap[key] = idx
				vtxBuffer = append(vtxBuffer, obj.Vertices[f.VtxIndices[i] - 1])
				normalBuffer = append(normalBuffer, obj.Normals[f.NormalIndices[i] - 1])
			}

			idxBuffer = append(idxBuffer, idx)
		}
	}

	//fmt.Printf("vtxBuffer = %v\n", vtxBuffer)
	vtxVBO := buffers.MakeVBOFromVec3s(vtxBuffer)
	defer vtxVBO.Delete()

	//fmt.Printf("normalBuffer = %v\n", normalBuffer)
	normalVBO := buffers.MakeVBOFromVec3s(normalBuffer)
	defer normalVBO.Delete()

	//fmt.Printf("idxBuffer = %v\n", idxBuffer)
	vao := buffers.MakeVAO(gl.TRIANGLES, len(idxBuffer))

	vao.AttachVBO(0, vtxVBO)
	vao.AttachVBO(1, normalVBO)
	vao.SetIndexBuffer(idxBuffer)

	return vao
}

func LoadOBJ(path string, diffuseColor *vmath.Vector4) *Object {
	objData := parseOBJ(path)

	bb := util.MakeBBox()
	for _,v := range objData.Vertices {
		bb.AddPoint(&v)
	}

	// fmt.Printf("bbox: %v\n", bb)


	vao := makeVAOFromOBJ(objData)
	return MakeObject(vao, bb, diffuseColor)
}

