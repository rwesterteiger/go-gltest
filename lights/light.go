package lights

import (
	"github.com/rwesterteiger/go-gltest/gbuffer"
	vmath "github.com/rwesterteiger/vectormath"
)

type Light interface {
	Delete()

	NeedDepthPass() bool
	BeginDepthPass() (projMat, viewMat *vmath.Matrix4)
	EndDepthPass()

	Render(gbuf *gbuffer.GBuffer, projMat, viewMat *vmath.Matrix4)
}
