package main

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/store/goleveldb"
	"github.com/blevesearch/bleve/mapping"
	"github.com/gohugoio/hugo/deps"
	"github.com/spf13/afero"

	"github.com/gohugoio/hugo/hugolib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	jww "github.com/spf13/jwalterweatherman"
)

func init() {
	HugoIndexCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	hugoCmdV = HugoIndexCmd
}

// flags
var Verbose bool

var hugoCmdV *cobra.Command

var HugoIndexCmd = &cobra.Command{
	Use:   "hugoidx",
	Short: "hugoidx builds a search index of your site",
	Long:  `hugoidx is the main command, used to build a search index of your Hugo site.`,
	Run: func(cmd *cobra.Command, args []string) {
		InitializeConfig()
		buildindex()
	},
}

func InitializeConfig() {
	LoadDefaultSettings()
	if hugoCmdV.PersistentFlags().Lookup("verbose").Changed {
		viper.Set("Verbose", Verbose)
	}
	// commands.InitializeConfig()
}

func LoadDefaultSettings() {
	viper.SetDefault("IndexDir", "search.bleve")
}

func buildindex() {

	mm := afero.NewOsFs()
	cfg, _, err := hugolib.LoadConfig(hugolib.ConfigSourceDescriptor{Fs: mm, Filename: "config.toml"})
	if err != nil {
		panic(err)
	}
	depCfg, buildCfg := deps.DepsCfg{Cfg: cfg}, hugolib.BuildCfg{SkipRender: true}
	h, err := hugolib.NewHugoSites(depCfg)
	if err != nil {
		panic(err)
	}
	err = h.Build(buildCfg)
	if err != nil {
		panic(err)
	}
	site := h.Sites[0]

	// create/index open the index
	index, err := createOpenIndex(viper.GetString("IndexDir"))
	if err != nil {
		jww.FATAL.Printf("error creating/opening index: %v", err)
	}

	for _, p := range site.Pages() {
		// current work around for issue #2
		if len(p.Title()) <= 0 {
			continue
		}
		jww.INFO.Printf("params: %#v", p.Params)
		rpl := p.RelPermalink()
		jww.INFO.Printf("Indexing: %s as - %s (% x)", p.Title, rpl, rpl)
		pi := NewPageForIndex(p)
		p.Summary()

		err = index.Index(rpl, pi)
		if err != nil {
			jww.FATAL.Printf("error indexing: %v", err)
		}
	}

	err = index.Close()
	if err != nil {
		jww.FATAL.Printf("error closing index: %v", err)
	}
}

func createOpenIndex(path string) (bleve.Index, error) {
	index, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		jww.INFO.Println("Creating Index: ", path)
		indexMapping, err := buildIndexMapping()
		if err != nil {
			return nil, err
		}
		index, err = bleve.NewUsing(path, indexMapping, bleve.Config.DefaultIndexType, goleveldb.Name, nil)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		jww.INFO.Println("Opening Index: ", path)
	}
	return index, nil
}

func buildIndexMapping() (mapping.IndexMapping, error) {
	rv := bleve.NewIndexMapping()

	return rv, nil
}

func main() {
	HugoIndexCmd.Execute()
}
