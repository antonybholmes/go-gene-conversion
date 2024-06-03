package geneconv

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/antonybholmes/go-sys"
	"github.com/rs/zerolog/log"
)

const MOUSE_TO_HUMAN_EXACT_SQL = `SELECT human.gene_id, human.gene_symbol, human.aliases, human.entrez, human.refseq, human.ensembl
	FROM mouse_terms, conversion, human
 	WHERE LOWER(mouse_terms.term) = LOWER(?1) AND conversion.mouse_gene_id = mouse_terms.gene_id AND human.gene_id = conversion.human_gene_id`

const HUMAN_TO_MOUSE_EXACT_SQL = `SELECT mouse.gene_id, mouse.gene_id, mouse.gene_symbol, mouse.aliases, mouse.entrez, mouse.refseq, mouse.ensembl
	FROM human_terms, conversion, mouse
	WHERE LOWER(human_terms.term) = LOWER(?1) AND conversion.human_gene_id = human_terms.gene_id AND mouse.gene_id = conversion.mouse_gene_id`

const MOUSE_TO_HUMAN_SQL = `SELECT human.gene_id, human.gene_symbol, human.aliases, human.entrez, human.refseq, human.ensembl
	FROM mouse_terms, conversion, human
 	WHERE mouse_terms.term LIKE ?1 AND conversion.mouse_gene_id = mouse_terms.gene_id AND human.gene_id = conversion.human_gene_id`

const HUMAN_TO_MOUSE_SQL = `SELECT mouse.gene_id, mouse.gene_id, mouse.gene_symbol, mouse.aliases, mouse.entrez, mouse.refseq, mouse.ensembl
	FROM human_terms, conversion, mouse
	WHERE human_terms.term LIKE ?1 AND conversion.human_gene_id = human_terms.gene_id AND mouse.gene_id = conversion.mouse_gene_id`

const MOUSE_SQL = `SELECT mouse.gene_id, mouse.gene_symbol, mouse.aliases, mouse.entrez, mouse.refseq, mouse.ensembl
	FROM mouse_terms, mouse
 	WHERE mouse_terms.term LIKE ?1 AND mouse.gene_id = mouse_terms.gene_id`

const HUMAN_SQL = `SELECT human.gene_id, human.gene_symbol, human.aliases, human.entrez, human.refseq, human.ensembl
	FROM human_terms, human
 	WHERE human_terms.term LIKE ?1 AND human.gene_id = human_terms.gene_id`

const MOUSE_EXACT_SQL = `SELECT mouse.gene_id, mouse.gene_symbol, mouse.aliases, mouse.entrez, mouse.refseq, mouse.ensembl
	 FROM mouse_terms, mouse
	  WHERE LOWER(mouse_terms.term) = LOWER(?1) AND mouse.gene_id = mouse_terms.gene_id`

const HUMAN_EXACT_SQL = `SELECT human.gene_id, human.gene_symbol, human.aliases, human.entrez, human.refseq, human.ensembl
	 FROM human_terms, human
	  WHERE LOWER(human_terms.term) = LOWER(?1) AND human.gene_id = human_terms.gene_id`

const HUMAN_TAXONOMY_ID = 9606
const MOUSE_TAXONOMY_ID = 10090

const HUMAN_SPECIES = "human"
const MOUSE_SPECIES = "mouse"

const STATUS_FOUND = "found"
const STATUS_NOT_FOUND = "not found"

type Taxonomy struct {
	TaxId   uint   `json:"taxId"`
	Species string `json:"species"`
}

type BaseGene struct {
	Taxonomy
	Id string `json:"id"`
}

type Gene struct {
	BaseGene
	Symbol  string   `json:"symbol"`
	Aliases []string `json:"aliases"`
	Entrez  []uint64 `json:"entrez"`
	RefSeq  []string `json:"refseq"`
	Ensembl []string `json:"ensembl"`
}

type Conversion struct {
	Search string `json:"id"`
	Genes  []Gene `json:"genes"`
}

type ConversionResults struct {
	From        Taxonomy     `json:"from"`
	To          Taxonomy     `json:"to"`
	Conversions []Conversion `json:"conversions"`
}

type GeneResult struct {
	Id    string `json:"id"`
	Genes []Gene `json:"genes"`
}

type GeneConvDB struct {
	db *sql.DB
}

func NewGeneConvDB(file string) *GeneConvDB {
	db := sys.Must(sql.Open("sqlite3", file))

	return &GeneConvDB{db: db}
}

func (geneconvdb *GeneConvDB) Close() {
	geneconvdb.db.Close()
}

func (geneconvdb *GeneConvDB) Convert(search string, fromSpecies string, toSpecies string, exact bool) (Conversion, error) {
	var aliases string
	var entrez string
	var refseq string
	var ensembl string
	var sql string

	var ret Conversion

	ret.Search = search
	ret.Genes = make([]Gene, 0)

	fromSpecies = strings.ToLower(fromSpecies)

	if fromSpecies == HUMAN_SPECIES {

		if exact {
			sql = HUMAN_TO_MOUSE_EXACT_SQL
		} else {
			sql = HUMAN_TO_MOUSE_SQL
		}
	} else {

		if exact {
			sql = MOUSE_TO_HUMAN_EXACT_SQL
		} else {
			sql = MOUSE_TO_HUMAN_SQL
		}
	}

	if !exact {
		search = fmt.Sprintf("%%%s%%", search)
	}

	log.Debug().Msgf("conv %s %s", search, sql)

	rows, err := geneconvdb.db.Query(sql, search)

	if err != nil {
		return ret, err
	}

	defer rows.Close()

	for rows.Next() {
		var gene = Gene{}

		if fromSpecies == HUMAN_SPECIES {
			gene.TaxId = HUMAN_TAXONOMY_ID
			gene.Species = HUMAN_SPECIES
		} else {
			gene.TaxId = MOUSE_TAXONOMY_ID
			gene.Species = MOUSE_SPECIES
		}

		err := rows.Scan(&gene.Id,
			&gene.Symbol,
			&aliases,
			&entrez,
			&refseq,
			&ensembl)

		if err != nil {
			return ret, err
		}

		for _, e := range strings.Split(entrez, ",") {
			n, err := strconv.ParseUint(e, 10, 64)

			if err == nil {
				gene.Entrez = append(gene.Entrez, n)
			}
		}

		gene.Aliases = strings.Split(aliases, ",")
		gene.RefSeq = strings.Split(refseq, ",")
		gene.Ensembl = strings.Split(ensembl, ",")

		ret.Genes = append(ret.Genes, gene)
	}

	return ret, nil
}

func (geneconvdb *GeneConvDB) Gene(search string, species string, exact bool) ([]Gene, error) {
	var aliases string
	var entrez string
	var refseq string
	var ensembl string
	var sql string

	species = strings.ToLower(species)

	var ret = make([]Gene, 0)

	if species == HUMAN_SPECIES {
		if exact {
			sql = HUMAN_EXACT_SQL
		} else {
			sql = HUMAN_SQL
		}
	} else {
		if exact {
			sql = MOUSE_EXACT_SQL
		} else {
			sql = MOUSE_SQL
		}
	}

	if !exact {
		search = fmt.Sprintf("%%%s%%", search)
	}

	rows, err := geneconvdb.db.Query(sql, search)

	if err != nil {
		log.Debug().Msgf("gggg %v", err)
		return ret, err
	}

	defer rows.Close()

	for rows.Next() {
		var gene Gene

		if species == HUMAN_SPECIES {
			gene.TaxId = HUMAN_TAXONOMY_ID
			gene.Species = HUMAN_SPECIES

		} else {
			gene.TaxId = MOUSE_TAXONOMY_ID
			gene.Species = MOUSE_SPECIES

		}

		err := rows.Scan(&gene.Id,
			&gene.Symbol,
			&aliases,
			&entrez,
			&refseq,
			&ensembl)

		if err != nil {
			return ret, nil //fmt.Errorf("there was an error with the database query")
		}

		// convert entrez to numbers
		for _, e := range strings.Split(entrez, ",") {
			n, err := strconv.ParseUint(e, 10, 64)

			if err == nil {
				gene.Entrez = append(gene.Entrez, n)
			}
		}

		gene.Aliases = strings.Split(aliases, ",")
		gene.RefSeq = strings.Split(refseq, ",")
		gene.Ensembl = strings.Split(ensembl, ",")
		ret = append(ret, gene)
	}

	return ret, nil
}