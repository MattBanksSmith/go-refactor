package internal

// eventually pass config into the application to drive refactor.
// Not yet implemented - need abstraction to form and work out the contract
type config struct {
}

type configItem struct {
	nodeType  string //type e.g. funcDecl
	operation string //replace, add or replace, rename

}
