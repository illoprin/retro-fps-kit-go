package post

// post processing pipeline model

import (
	"fmt"

	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/camera"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/logger"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/rhi"
)

// PassDescriptor - describes pass and its dependencies
type PassDescriptor struct {
	ID    PassID
	Pass  PostProcessingPass
	After PassID // previous effect
}

// PostProcessContext - runtime execution context
type PostProcessContext struct {
	GBuffer   *rhi.Framebuffer
	Result    *pipeline.DeferredRenderResult
	Camera    *camera.Camera3D
	DeltaTime float32
}

// CameraProvider - minimal camera interface
type CameraProvider interface {
	GetProjection() [16]float32
}

// PostProcessingPipeline - main pipeline struct
type PostProcessingPipeline struct {
	descriptors map[PassID]*PassDescriptor

	// baked execution list
	execution []PostProcessingPass

	dirty bool
}

// NewPipeline - create pipeline
func NewPipeline() *PostProcessingPipeline {
	return &PostProcessingPipeline{
		descriptors: make(map[PassID]*PassDescriptor),
		dirty:       true,
	}
}

// AddPass - add new pass descriptor
func (p *PostProcessingPipeline) AddPass(desc *PassDescriptor) {

	if len(p.descriptors) == 0 {
		p.descriptors[desc.ID] = desc
		p.dirty = true
		return
	}

	if _, ok := p.descriptors[desc.ID]; ok {
		logger.Warnf("id repeating")
		return
	}

	nextOne := p.getNext(desc.After)
	if n, ok := p.descriptors[nextOne]; ok {
		n.After = desc.ID
	}
	p.descriptors[desc.ID] = desc

}

// getNext - returns next node
func (p *PostProcessingPipeline) getNext(id PassID) PassID {

	for _, node := range p.descriptors {
		if node.After == id {
			return node.ID
		}
	}

	return ""
}

// RemovePass - removes pass
func (p *PostProcessingPipeline) RemovePass(id PassID) {
	d, ok := p.descriptors[id]

	if !ok {
		return
	}

	nextOne := p.getNext(id)
	previousOne := d.After

	if nextD, ok := p.descriptors[nextOne]; ok {
		nextD.After = previousOne
	}

	if d.Pass != nil {
		d.Pass.Delete()
	}

	delete(p.descriptors, id)
	p.dirty = true
}

// ReplacePass - replace existing pass
func (p *PostProcessingPipeline) ReplacePass(id PassID, newPass PostProcessingPass) {
	if d, ok := p.descriptors[id]; ok {
		d.Pass = newPass
		p.dirty = true
	}
}

// GetPass - generic typed pass access
func (p *PostProcessingPipeline) GetPass(id PassID) PostProcessingPass {
	if d, ok := p.descriptors[id]; ok {
		return d.Pass
	}
	return nil
}

// GetConfigTyped - generic typed config access
func GetConfigTyped[T any](p *PostProcessingPipeline, id PassID) *T {
	pass := p.GetPass(id)
	if pass == nil {
		return nil
	}

	cfg, ok := pass.GetConfig().(*T)
	if !ok {
		return nil
	}

	return cfg
}

func (p *PostProcessingPipeline) ResolveOrder() ([]PassID, error) {
	// indexing
	nextMap := make(map[PassID]PassID)
	var root PassID

	// build next map (map[next_node_id] = id)
	for id, node := range p.descriptors {
		if node.After == "" {
			root = id
		} else {
			nextMap[node.After] = id
		}
	}

	// resolve empty root
	if root == "" {
		logger.Errorf("root not found")
		return nil, fmt.Errorf("root not found")
	}

	// build list
	list := make([]PassID, 0, len(p.descriptors))
	current := root

	for current != "" {
		list = append(list, current)
		current = nextMap[current]
	}

	if len(list) != len(p.descriptors) {
		return nil, fmt.Errorf("pipeline is broken or has branches/cycles")
	}

	logger.Infof("passes order resolved %v", list)

	return list, nil
}

// Build - bakes chaotic descriptors list to ordered execution list
func (p *PostProcessingPipeline) Build() error {

	if len(p.descriptors) <= 0 {
		return nil
	}

	if !p.dirty {
		return nil
	}

	l, err := p.ResolveOrder()
	if err != nil {
		return err
	}

	p.execution = p.execution[:0]

	for _, id := range l {
		p.execution = append(p.execution, p.descriptors[id].Pass)
	}

	logger.Infof("post processing pass built successfully")

	p.dirty = false

	return nil
}

// GetExecutionList returns current execution list
// UNSAFE! may not be relevant
func (p *PostProcessingPipeline) GetExecutionList() []PostProcessingPass {
	if p.dirty {
		logger.Warnf("GetExecutionList: execution list is dirty, Build() required")
	}
	return p.execution
}

// Execute - run baked pipeline
// and returns result texture and framebuffer
func (p *PostProcessingPipeline) Execute(ctx PostProcessContext) (resultColor *rhi.Texture, resultFbo *rhi.Framebuffer) {

	if p.dirty {
		logger.Warnf("dirty pipeline")
	}

	res := ctx.Result
	resultFbo = ctx.GBuffer
	resultColor = res.Color

	for _, pass := range p.execution {

		if !pass.Use() {
			continue
		}

		if h, ok := pass.(HasProjection); ok {
			proj := ctx.Camera.Projection
			h.SetProjectionMatrix(proj)
		}

		if h, ok := pass.(HasDeltaTime); ok {
			h.SetDeltaTime(ctx.DeltaTime)
		}

		pass.RenderPass(res)

		res.Color = pass.GetColor()
		resultFbo = pass.GetResultFramebuffer()
	}

	return resultColor, resultFbo
}

func (p *PostProcessingPipeline) ResizeCallback() {
	for _, n := range p.execution {
		n.ResizeCallback()
	}
}

func (p *PostProcessingPipeline) Delete() {
	for _, n := range p.execution {
		if n != nil {
			n.Delete()
		}
	}

	p.dirty = true
}
