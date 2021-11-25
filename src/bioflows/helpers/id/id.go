package id

import "github.com/aidarkhanov/nanoid"

const (
	FLOWS_ALPHABET = nanoid.DefaultAlphabet
	SIZE = 15
)

func NewID() (string,error){

	return nanoid.Generate(FLOWS_ALPHABET,SIZE)
}