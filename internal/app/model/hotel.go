package model

type Hotel struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Address     *Address `json:"address"`
	StarsCount  int      `json:"stars_count"`
	Description string   `json:"description"`
}

// TODO СДЕЛАТЬ БЛЯДКСУЮ ВАЛИДАЦИЮ
