package keqing

var (
	skill = []float64{
		0.504,
		0.5418,
		0.5796,
		0.63,
		0.6678,
		0.7056,
		0.756,
		0.8064,
		0.8568,
		0.9072,
		0.9576,
		1.008,
		1.071,
		1.134,
		1.197,
	}
	skillPress = []float64{
		1.68,
		1.806,
		1.932,
		2.1,
		2.226,
		2.352,
		2.52,
		2.688,
		2.856,
		3.024,
		3.192,
		3.36,
		3.57,
		3.78,
		3.99,
	}
	skillCA = []float64{
		0.84,
		0.903,
		0.966,
		1.05,
		1.113,
		1.176,
		1.26,
		1.344,
		1.428,
		1.512,
		1.596,
		1.68,
		1.785,
		1.89,
		1.995,
	}
	burstInitial = []float64{
		0.88,
		0.946,
		1.012,
		1.1,
		1.166,
		1.232,
		1.32,
		1.408,
		1.496,
		1.584,
		1.672,
		1.76,
		1.87,
		1.98,
		2.09,
	}
	burstDot = []float64{
		0.24,
		0.258,
		0.276,
		0.3,
		0.318,
		0.336,
		0.36,
		0.384,
		0.408,
		0.432,
		0.456,
		0.48,
		0.51,
		0.54,
		0.57,
	}
	burstFinal = []float64{
		1.888,
		2.0296,
		2.1712,
		2.36,
		2.5016,
		2.6432,
		2.832,
		3.0208,
		3.2096,
		3.3984,
		3.5872,
		3.776,
		4.012,
		4.248,
		4.484,
	}
	attack = [][][]float64{
		{ //1
			{
				0.4102,
				0.4436,
				0.477,
				0.5247,
				0.5581,
				0.5962,
				0.6487,
				0.7012,
				0.7537,
				0.8109,
				0.8681,
				0.9254,
				0.9826,
				1.0399,
				1.0971,
			},
		},
		{ //2
			{
				0.4102,
				0.4436,
				0.477,
				0.5247,
				0.5581,
				0.5962,
				0.6487,
				0.7012,
				0.7537,
				0.8109,
				0.8681,
				0.9254,
				0.9826,
				1.0399,
				1.0971,
			},
		},
		{ //3
			{
				0.5444,
				0.5887,
				0.633,
				0.6963,
				0.7406,
				0.7913,
				0.8609,
				0.9305,
				1.0001,
				1.0761,
				1.1521,
				1.228,
				1.304,
				1.3799,
				1.4559,
			},
		},
		{ //4
			{
				0.3148,
				0.3404,
				0.366,
				0.4026,
				0.4282,
				0.4575,
				0.4978,
				0.538,
				0.5783,
				0.6222,
				0.6661,
				0.71,
				0.754,
				0.7979,
				0.8418,
			},
			{
				0.344,
				0.372,
				0.4,
				0.44,
				0.468,
				0.5,
				0.544,
				0.588,
				0.632,
				0.68,
				0.728,
				0.776,
				0.824,
				0.872,
				0.92,
			},
		},
		{ //5
			{
				0.6699,
				0.7245,
				0.779,
				0.8569,
				0.9114,
				0.9738,
				1.0594,
				1.1451,
				1.2308,
				1.3243,
				1.4178,
				1.5113,
				1.6047,
				1.6982,
				1.7917,
			},
		},
	}
)
