rm data/modules/geneconv/geneconv.db

python step1_conv_table.py
python step2_human.py
python step3_mouse.py

cat tables.sql | sqlite3 data/modules/geneconv/geneconv.db
cat data/modules/geneconv/conversion.sql | sqlite3 data/modules/geneconv/geneconv.db
cat data/modules/geneconv/human.sql | sqlite3 data/modules/geneconv/geneconv.db
cat data/modules/geneconv/mouse.sql | sqlite3 data/modules/geneconv/geneconv.db
cat data/modules/geneconv/human_terms.sql | sqlite3 data/modules/geneconv/geneconv.db
cat data/modules/geneconv/mouse_terms.sql | sqlite3 data/modules/geneconv/geneconv.db
