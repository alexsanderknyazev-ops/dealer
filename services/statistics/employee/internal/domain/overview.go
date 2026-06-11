package domain

type DealStageCount struct {
	Stage string
	Count int64
}

type Overview struct {
	CustomersCount    int64
	VehiclesCount     int64
	DealsCount        int64
	DealsByStage      []DealStageCount
	TotalRevenue      float64
	PartsCount        int64
	DealerPointsCount int64
}
