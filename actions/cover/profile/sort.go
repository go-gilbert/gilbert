package profile

const (
	// ByName means sort by name
	ByName = "name"

	// ByCoverage means sort by coverage percentage
	ByCoverage = "coverage"
)

type sortSelector func(asc bool, k1, k2 string) bool

func pkgByPercentage(p Packages) sortSelector {
	return func(asc bool, pkg1, pkg2 string) bool {
		pkg1cover, pkg2cover := p[pkg1].Percentage(), p[pkg2].Percentage()

		if asc {
			return pkg1cover < pkg2cover
		}

		return pkg1cover > pkg2cover
	}
}

func byName(asc bool, pkg1, pkg2 string) bool {
	if asc {
		return pkg1 < pkg2
	}

	return pkg1 > pkg2
}

type packageSorter struct {
	asc  bool
	keys []string
	by   sortSelector
}

func (p *packageSorter) Len() int {
	return len(p.keys)
}

func (p *packageSorter) Swap(i, j int) {
	p.keys[i], p.keys[j] = p.keys[j], p.keys[i]
}

func (p *packageSorter) Less(i, j int) bool {
	return p.by(p.asc, p.keys[i], p.keys[j])
}
