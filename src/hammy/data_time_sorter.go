package hammy

// Struct for sorting []IncomingValueData slice by Timestamp
type DataTimeSorter struct {
	Data *[]IncomingValueData
}

func (ds *DataTimeSorter) Len() int {
	return len(*ds.Data)
}

func (ds *DataTimeSorter) Less(i, j int) bool {
	return (*ds.Data)[i].Timestamp < (*ds.Data)[j].Timestamp
}

func (ds *DataTimeSorter) Swap(i, j int) {
	(*ds.Data)[i], (*ds.Data)[j] = (*ds.Data)[j], (*ds.Data)[i]
}