package math

import mgl "github.com/go-gl/mathgl/mgl32"

// determines the order in which transformations are applied
type TransformOrder int

const (
	OrderSRT TransformOrder = iota // Scale -> Rotate -> Translate (by default)
	OrderSTR                       // Scale -> Translate -> Rotate
	OrderRST                       // Rotate -> Scale -> Translate
	OrderRTS                       // Rotate -> Translate -> Scale
	OrderTSR                       // Translate -> Scale -> Rotate
	OrderTRS                       // Translate -> Rotate -> Scale
)

// GetTransformMatrixWithOrder returns transformation matrix in specified order
func GetTransformMatrixWithOrder(order TransformOrder, pos mgl.Vec3, rot mgl.Vec3, scl mgl.Vec3, pvt mgl.Vec3) mgl.Mat4 {
	// Начинаем с единичной матрицы
	model := mgl.Ident4()

	// scaling
	scaleMat := mgl.Scale3D(scl.X(), scl.Y(), scl.Z())

	// get combined rotation matrix
	rotMat := getRotationMatrix(rot)

	// translation matrix
	transMat := mgl.Translate3D(pos.X(), pos.Y(), pos.Z())

	// pivot translation matrix
	pivotMat := mgl.Translate3D(pvt.X(), pvt.Y(), pvt.Z())
	// invPivotMat := mgl32.Translate3D(-pvt.X(), -pvt.Y(), -pvt.Z())

	// apply transformations in specific order
	switch order {
	case OrderSRT:
		// Scale -> Rotate -> Translate
		model = model.Mul4(scaleMat)
		model = model.Mul4(rotMat)
		model = model.Mul4(transMat)

	case OrderSTR:
		// Scale -> Translate -> Rotate
		model = model.Mul4(scaleMat)
		model = model.Mul4(transMat)
		model = model.Mul4(rotMat)

	case OrderRST:
		// Rotate -> Scale -> Translate
		model = model.Mul4(rotMat)
		model = model.Mul4(scaleMat)
		model = model.Mul4(transMat)

	case OrderRTS:
		// Rotate -> Translate -> Scale
		model = model.Mul4(rotMat)
		model = model.Mul4(transMat)
		model = model.Mul4(scaleMat)

	case OrderTSR:
		// Translate -> Scale -> Rotate
		model = model.Mul4(transMat)
		model = model.Mul4(scaleMat)
		model = model.Mul4(rotMat)

	case OrderTRS:
		// Translate -> Rotate -> Scale
		model = model.Mul4(transMat)
		model = model.Mul4(rotMat)
		model = model.Mul4(scaleMat)

	default:
		// by default SRT
		model = model.Mul4(scaleMat)
		model = model.Mul4(rotMat)
		model = model.Mul4(transMat)
	}

	// apply pivot (if needed)
	if pvt != (mgl.Vec3{}) {
		model = pivotMat.Mul4(model)
	}

	return model
}

// getRotationMatrix returns combined rotation matrix
func getRotationMatrix(rot mgl.Vec3) mgl.Mat4 {
	// deg to rad
	rotX := mgl.DegToRad(rot.X())
	rotY := mgl.DegToRad(rot.Y())
	rotZ := mgl.DegToRad(rot.Z())

	// create matrix for each axis
	matRotX := mgl.HomogRotate3DX(rotX)
	matRotY := mgl.HomogRotate3DY(rotY)
	matRotZ := mgl.HomogRotate3DZ(rotZ)

	// default order: YXZ (Yaw -> Pitch -> Roll)
	return matRotZ.Mul4(matRotY).Mul4(matRotX)
}
