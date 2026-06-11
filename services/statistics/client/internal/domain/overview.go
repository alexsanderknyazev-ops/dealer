package domain

type ReviewStatusCount struct {
	Status string
	Count  int64
}

type Overview struct {
	ClientsCount         int64
	ClientVehiclesCount  int64
	RegisteredUsersCount int64
	ReviewsCount         int64
	AverageRating        float64
	ReviewsByStatus      []ReviewStatusCount
}
