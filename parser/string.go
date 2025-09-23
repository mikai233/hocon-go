package parser

var ForbiddenTable = [256]bool{
	'$':  true,
	'"':  true,
	'{':  true,
	'}':  true,
	'[':  true,
	']':  true,
	':':  true,
	'=':  true,
	',':  true,
	'+':  true,
	'#':  true,
	'`':  true,
	'^':  true,
	'?':  true,
	'!':  true,
	'@':  true,
	'*':  true,
	'&':  true,
	'\\': true,
}
