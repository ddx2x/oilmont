package micro

// add deployment to kubernetes registry plugins
//import _ "github.com/micro/go-plugins/registry/kubernetes"

type IMicroServer interface {
	Run() error
	UUID() string
}
