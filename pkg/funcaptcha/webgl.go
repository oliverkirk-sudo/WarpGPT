package funcaptcha

import "strings"

const (
	webglExtensions             = "ANGLE_instanced_arrays;EXT_blend_minmax;EXT_color_buffer_half_float;EXT_disjoint_timer_query;EXT_float_blend;EXT_frag_depth;EXT_shader_texture_lod;EXT_texture_compression_bptc;EXT_texture_compression_rgtc;EXT_texture_filter_anisotropic;EXT_sRGB;KHR_parallel_shader_compile;OES_element_index_uint;OES_fbo_render_mipmap;OES_standard_derivatives;OES_texture_float;OES_texture_float_linear;OES_texture_half_float;OES_texture_half_float_linear;OES_vertex_array_object;WEBGL_color_buffer_float;WEBGL_compressed_texture_s3tc;WEBGL_compressed_texture_s3tc_srgb;WEBGL_debug_renderer_info;WEBGL_debug_shaders;WEBGL_depth_texture;WEBGL_draw_buffers;WEBGL_lose_context;WEBGL_multi_draw" // this.getWebGLKeys();
	webglRenderer               = "WebKit WebGL"
	webglVendor                 = "WebKit"
	webglVersion                = "WebGL 1.0 (OpenGL ES 2.0 Chromium)"
	webglShadingLanguageVersion = "WebGL GLSL ES 1.0 (OpenGL ES GLSL ES 1.0 Chromium)"
	webglAliasedLineWidthRange  = "[1, 10]"
	webglAliasedPointSizeRange  = "[1, 2047]"
	webglAntialiasing           = "yes"
	webglBits                   = "8,8,24,8,8,0"
	webglMaxParams              = "16,64,32768,1024,32768,32,32768,31,16,32,1024"
	webglMaxViewportDims        = "[32768, 32768]"
	webglUnmaskedVendor         = "Google Inc. (NVIDIA Corporation)"
	webglUnmaskedRenderer       = "ANGLE (NVIDIA Corporation, NVIDIA GeForce RTX 3060 Ti/PCIe/SSE2, OpenGL 4.5.0)"
	webglFsfParams              = "23,127,127,10,15,15,10,15,15"
	webglFsiParams              = "0,31,30,0,31,30,0,31,30"
	webglVsfParams              = "23,127,127,10,15,15,10,15,15"
	webglVsiParams              = "0,31,30,0,31,30,0,31,30"
)

var (
	webglExtensionsHash = getWebglExtensionsHash()
)

func getWebglExtensionsHash() string {
	return x64hash128(webglExtensions, 0)
}

func getWebglHashWebgl() string {
	//aZ['webgl_hash' + cr(f_a_gY.X)] = this['x64hash128'](aC(aZ, function(b3) {
	//	return b3;
	//})[cr(f_a_gY.Y)](','));

	var webglList []string
	webglList = append(webglList, webglExtensions)
	webglList = append(webglList, webglExtensionsHash)
	webglList = append(webglList, webglRenderer)
	webglList = append(webglList, webglVendor)
	webglList = append(webglList, webglVersion)
	webglList = append(webglList, webglShadingLanguageVersion)
	webglList = append(webglList, webglAliasedLineWidthRange)
	webglList = append(webglList, webglAliasedPointSizeRange)
	webglList = append(webglList, webglAntialiasing)
	webglList = append(webglList, webglBits)
	webglList = append(webglList, webglMaxParams)
	webglList = append(webglList, webglMaxViewportDims)
	webglList = append(webglList, webglUnmaskedVendor)
	webglList = append(webglList, webglUnmaskedRenderer)
	webglList = append(webglList, webglFsfParams)
	webglList = append(webglList, webglFsiParams)
	webglList = append(webglList, webglVsfParams)
	webglList = append(webglList, webglVsiParams)
	return x64hash128(strings.Join(webglList, ","), 0)
}
