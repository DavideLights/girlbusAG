package assets

import _ "embed"

// PostProcessShaderSrc contains the embedded Kage post-processing shader source
// (Warm CRT, VHS Noir, and Pixel Crisp modes with Chronos & Rewind cues).
//
//go:embed shaders/postprocess.kage
var PostProcessShaderSrc []byte
