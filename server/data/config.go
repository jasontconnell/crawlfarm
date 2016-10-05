package data

type Headers map[string]string

type Config struct {
    Listen string `json:"listen"`
    Root string                 `json:"root"`
    VirtualPaths []string       `json:"virtualPaths"`
    Headers Headers             `json:"headers"`
    ErrorFileName string            `json:"errorFile"`
}