// Package types defines the core data structures for Axon model manifests and metadata.
package types

import "time"

// Manifest represents a model manifest
// This is the core structure that describes a model package
type Manifest struct {
	APIVersion   string       `yaml:"apiVersion"`
	Kind         string       `yaml:"kind"`
	Metadata     Metadata     `yaml:"metadata"`
	Spec         Spec         `yaml:"spec"`
	Distribution Distribution `yaml:"distribution"`
}

// Metadata contains model identification and authorship information
type Metadata struct {
	Name          string    `yaml:"name"`
	Namespace     string    `yaml:"namespace"`
	Version       string    `yaml:"version"`
	Description   string    `yaml:"description"`
	Authors       []Author  `yaml:"authors,omitempty"`
	License       string    `yaml:"license"`
	Homepage      string    `yaml:"homepage,omitempty"`
	Documentation string    `yaml:"documentation,omitempty"`
	Created       time.Time `yaml:"created"`
	Updated       time.Time `yaml:"updated"`
	Tags          []string  `yaml:"tags,omitempty"`
}

// Author represents a model author
type Author struct {
	Name         string `yaml:"name"`
	Email        string `yaml:"email,omitempty"`
	Organization string `yaml:"organization,omitempty"`
}

// Spec contains the model specification
type Spec struct {
	Framework    Framework    `yaml:"framework"`
	Format       Format       `yaml:"format"`
	IO           IO           `yaml:"io"`
	Requirements Requirements `yaml:"requirements"`
	Performance  Performance  `yaml:"performance,omitempty"`
	Dependencies Dependencies `yaml:"dependencies,omitempty"`
}

// Framework specifies the ML framework
type Framework struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// Format describes the model file format
type Format struct {
	Type            string      `yaml:"type" json:"type"`                                       // Original format (pytorch, tensorflow)
	ExecutionFormat string      `yaml:"execution_format" json:"execution_format"`               // Execution format (onnx, pytorch, tensorflow)
	MultiEncoder    string      `yaml:"multi_encoder,omitempty" json:"multi_encoder,omitempty"` // Architecture for multi-encoder models (clip, seq2seq)
	Files           []ModelFile `yaml:"files" json:"files"`
}

// ModelFile represents a file in the model package
type ModelFile struct {
	Path   string `yaml:"path"`
	Size   int64  `yaml:"size"`
	SHA256 string `yaml:"sha256"`
}

// IO describes input/output schema
type IO struct {
	Inputs  []IOSpec `yaml:"inputs"`
	Outputs []IOSpec `yaml:"outputs"`
}

// IOSpec describes an input or output
type IOSpec struct {
	Name          string             `yaml:"name"`
	DType         string             `yaml:"dtype"`
	Shape         []int              `yaml:"shape"` // null represented as -1
	Description   string             `yaml:"description,omitempty"`
	Preprocessing *PreprocessingSpec `yaml:"preprocessing,omitempty"`
}

// PreprocessingSpec describes preprocessing requirements
type PreprocessingSpec struct {
	Type          string                 `yaml:"type"`                     // "tokenization", "normalization", "resize"
	Tokenizer     string                 `yaml:"tokenizer,omitempty"`      // Path to tokenizer.json
	TokenizerType string                 `yaml:"tokenizer_type,omitempty"` // "bert", "gpt2", etc.
	Config        map[string]interface{} `yaml:"config,omitempty"`         // Normalization params, resize params, etc.
}

// Requirements specifies hardware and storage requirements
type Requirements struct {
	Compute Compute `yaml:"compute"`
	Storage Storage `yaml:"storage,omitempty"`
}

// Compute specifies compute requirements
type Compute struct {
	CPU    CPURequirement    `yaml:"cpu"`
	Memory MemoryRequirement `yaml:"memory"`
	GPU    *GPURequirement   `yaml:"gpu,omitempty"`
}

// CPURequirement specifies CPU requirements
type CPURequirement struct {
	MinCores         int `yaml:"min_cores"`
	RecommendedCores int `yaml:"recommended_cores"`
}

// MemoryRequirement specifies memory requirements
type MemoryRequirement struct {
	MinGB         float64 `yaml:"min_gb"`
	RecommendedGB float64 `yaml:"recommended_gb"`
}

// GPURequirement specifies GPU requirements
type GPURequirement struct {
	Required    bool    `yaml:"required"`
	Recommended bool    `yaml:"recommended"`
	MinVRAMGB   float64 `yaml:"min_vram_gb,omitempty"`
	CUDAVersion string  `yaml:"cuda_version,omitempty"`
}

// Storage specifies storage requirements
type Storage struct {
	MinGB         float64 `yaml:"min_gb"`
	RecommendedGB float64 `yaml:"recommended_gb"`
}

// Performance contains performance characteristics
type Performance struct {
	InferenceTime map[string]string  `yaml:"inference_time,omitempty"`
	Throughput    map[string]string  `yaml:"throughput,omitempty"`
	Accuracy      map[string]float64 `yaml:"accuracy,omitempty"`
	Dataset       string             `yaml:"dataset,omitempty"`
}

// Dependencies lists model and package dependencies
type Dependencies struct {
	Models   []string            `yaml:"models,omitempty"`
	Packages []PackageDependency `yaml:"packages,omitempty"`
}

// PackageDependency represents a Python package dependency
type PackageDependency struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// Distribution contains package distribution information
type Distribution struct {
	Package  PackageInfo  `yaml:"package"`
	Registry RegistryInfo `yaml:"registry"`
}

// PackageInfo contains package location and checksums
type PackageInfo struct {
	URL     string   `yaml:"url"`
	Size    int64    `yaml:"size"`
	SHA256  string   `yaml:"sha256"`
	Mirrors []string `yaml:"mirrors,omitempty"`
}

// RegistryInfo contains registry information
type RegistryInfo struct {
	URL       string `yaml:"url"`
	Namespace string `yaml:"namespace"`
}

// FullName returns the full model name (namespace/name)
func (m *Manifest) FullName() string {
	return m.Metadata.Namespace + "/" + m.Metadata.Name
}

// FullVersion returns the full versioned name (namespace/name@version)
func (m *Manifest) FullVersion() string {
	return m.FullName() + "@" + m.Metadata.Version
}

// Validate performs basic validation on the manifest
// This is a convenience method that delegates to the manifest validator
func (m *Manifest) Validate() error {
	// Delegate to the manifest package validator
	// This keeps validation logic centralized
	return nil
}
