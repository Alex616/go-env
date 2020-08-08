// Package env parses environments using the fields from a struct.
//
// For example,
//
//	var envs struct {
//		Iter int
//		Debug bool
//	}
//	env.MustParse(&envs)
//
// defines two environments, which can be set using any of
//
//	iter=1 debug=true ./example  // debug is a boolean flag so its value is set to true
package env
