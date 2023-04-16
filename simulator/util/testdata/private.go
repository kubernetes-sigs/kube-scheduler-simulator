package testdata

type privateStruct2 struct {
	privateField string
}

type privateStruct struct {
	privateField  string
	privateField2 privateStruct2
}

func New(msg string) *privateStruct {
	return &privateStruct{privateField: msg, privateField2: privateStruct2{privateField: msg}}
}
