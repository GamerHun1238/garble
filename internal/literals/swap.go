package literals

import (
	"go/ast"
	"go/token"
	"math"
	mathrand "math/rand"
)

type swap struct{}

// check that the obfuscator interface is implemented
var _ obfuscator = swap{}

func getIndexType(dataLen int) string {
	switch {
	case dataLen <= math.MaxUint8:
		return "byte"
	case dataLen <= math.MaxUint16:
		return "uint16"
	case dataLen <= math.MaxUint32:
		return "uint32"
	default:
		return "uint64"
	}
}

func positionsToSlice(data []int) *ast.CompositeLit {
	arr := &ast.CompositeLit{
		Type: &ast.ArrayType{
			Len: &ast.Ellipsis{}, // Performance optimization
			Elt: ident(getIndexType(len(data))),
		},
		Elts: []ast.Expr{},
	}
	for _, data := range data {
		arr.Elts = append(arr.Elts, intLiteral(data))
	}
	return arr
}

// Generates a random even swap count based on the length of data
func generateSwapCount(dataLen int) int {
	swapCount := dataLen

	maxExtraPositions := dataLen / 2 // Limit the number of extra positions to half the data length
	if maxExtraPositions > 1 {
		swapCount += mathrand.Intn(maxExtraPositions)
	}
	if swapCount%2 != 0 { // Swap count must be even
		swapCount++
	}
	return swapCount
}

func (x swap) obfuscate(data []byte) *ast.BlockStmt {
	swapCount := generateSwapCount(len(data))
	shiftKey := byte(mathrand.Intn(math.MaxUint8))

	positions := genRandIntSlice(len(data), swapCount)
	for i := len(positions) - 2; i >= 0; i -= 2 {
		// Generate local key for xor based on random key and byte position
		localKey := byte(i) + byte(positions[i]^positions[i+1]) + shiftKey
		// Swap bytes from i+1 to i and xor using local key
		data[positions[i]], data[positions[i+1]] = data[positions[i+1]]^localKey, data[positions[i]]^localKey
	}

	return &ast.BlockStmt{List: []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ident("data")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{dataToByteSlice(data)},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{ident("positions")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{positionsToSlice(positions)},
		},
		&ast.ForStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{ident("i")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{intLiteral(0)},
			},
			Cond: &ast.BinaryExpr{
				X:  ident("i"),
				Op: token.LSS,
				Y:  intLiteral(len(positions)),
			},
			Post: &ast.AssignStmt{
				Lhs: []ast.Expr{ident("i")},
				Tok: token.ADD_ASSIGN,
				Rhs: []ast.Expr{intLiteral(2)},
			},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ident("localKey")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.BinaryExpr{
						X: &ast.BinaryExpr{
							X:  callExpr(ident("byte"), ident("i")),
							Op: token.ADD,
							Y: callExpr(ident("byte"), &ast.BinaryExpr{
								X:  indexExpr("positions", ident("i")),
								Op: token.XOR,
								Y: indexExpr("positions", &ast.BinaryExpr{
									X:  ident("i"),
									Op: token.ADD,
									Y:  intLiteral(1),
								}),
							}),
						},
						Op: token.ADD,
						Y:  intLiteral(int(shiftKey)),
					}},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						indexExpr("data", indexExpr("positions", ident("i"))),
						indexExpr("data", indexExpr("positions", &ast.BinaryExpr{
							X:  ident("i"),
							Op: token.ADD,
							Y:  intLiteral(1),
						})),
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.BinaryExpr{
							X: indexExpr("data", indexExpr("positions", &ast.BinaryExpr{
								X:  ident("i"),
								Op: token.ADD,
								Y:  intLiteral(1),
							})),
							Op: token.XOR,
							Y:  ident("localKey"),
						},
						&ast.BinaryExpr{
							X:  indexExpr("data", indexExpr("positions", ident("i"))),
							Op: token.XOR,
							Y:  ident("localKey"),
						},
					},
				},
			}},
		},
	}}
}
