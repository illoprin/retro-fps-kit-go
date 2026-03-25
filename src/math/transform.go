package math

import "github.com/go-gl/mathgl/mgl32"

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
func GetTransformMatrixWithOrder(order TransformOrder, pos mgl32.Vec3, rot mgl32.Vec3, scl mgl32.Vec3, pvt mgl32.Vec3) mgl32.Mat4 {
	// Начинаем с единичной матрицы
	model := mgl32.Ident4()

	// scaling
	scaleMat := mgl32.Scale3D(scl.X(), scl.Y(), scl.Z())

	// get combined rotation matrix
	rotMat := getRotationMatrix(rot)

	// translation matrix
	transMat := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())

	// pivot translation matrix
	pivotMat := mgl32.Translate3D(pvt.X(), pvt.Y(), pvt.Z())
	invPivotMat := mgl32.Translate3D(-pvt.X(), -pvt.Y(), -pvt.Z())

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
	if pvt != (mgl32.Vec3{}) {
		model = invPivotMat.Mul4(model).Mul4(pivotMat)
	}

	return model
}

// getRotationMatrix returns combined rotation matrix
func getRotationMatrix(rot mgl32.Vec3) mgl32.Mat4 {
	// deg to rad
	rotX := mgl32.DegToRad(rot.X())
	rotY := mgl32.DegToRad(rot.Y())
	rotZ := mgl32.DegToRad(rot.Z())

	// create matrix for each axis
	matRotX := mgl32.HomogRotate3DX(rotX)
	matRotY := mgl32.HomogRotate3DY(rotY)
	matRotZ := mgl32.HomogRotate3DZ(rotZ)

	// default order: YXZ (Yaw -> Pitch -> Roll)
	return matRotZ.Mul4(matRotY).Mul4(matRotX)
}
