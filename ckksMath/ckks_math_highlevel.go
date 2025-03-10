package ckksMath

import (
	"errors"
)

// ArraySum Returns the encrypted sum of all elements of passed array in []byte
func ArraySum(encryptedDataArray [][]byte) ([]byte, error) {
	if len(encryptedDataArray) == 0 {
		return nil, errors.New("cannot use empty array")
	}

	sumCiphertext, err := MakeZeroCiphertext(encryptedDataArray[0])
	if err != nil {
		return nil, err
	}

	for _, encryptedData := range encryptedDataArray {
		ciphertext, err := unmarshallIntoNewCiphertext(encryptedData)
		if err != nil {
			return nil, err
		}

		CkksEvaluator.Add(sumCiphertext, ciphertext, sumCiphertext)
	}

	return sumCiphertext.MarshalBinary()
}

// ArrayMean Calculates the encrypted mean of all elements of passed array in []byte
func ArrayMean(encryptedDataArray [][]byte) ([]byte, error) {
	sum, err := ArraySum(encryptedDataArray)
	if err != nil {
		return nil, err
	}

	ciphertext, err := unmarshallIntoNewCiphertext(sum)
	if err != nil {
		return nil, err
	}

	return CkksEvaluator.MultByConstNew(ciphertext, 1.0/float64(len(encryptedDataArray))).MarshalBinary()
}

// MovingAverage Returns an array, containing len(encryptedDataArray) - windowSize elements,
// each representing a calculated mean of numbers within a shifting window of size windowSize
func MovingAverage(encryptedDataArray [][]byte, windowSize int) ([][]byte, error) {
	movingArrayLen := len(encryptedDataArray) - windowSize
	r := make([][]byte, movingArrayLen)
	for i := 0; i < movingArrayLen; i++ {
		var err error
		r[i], err = ArrayMean(encryptedDataArray[i : i+windowSize])
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

// дисперсия
func Variance(encryptedDataArray [][]byte) ([]byte, error) {
	if len(encryptedDataArray) == 0 {
		return nil, errors.New("cannot use empty array")
	}

	mean, err := ArrayMean(encryptedDataArray)
	if err != nil {
		return nil, err
	}

	ciphertextSum, err := MakeZeroCiphertext(encryptedDataArray[0])
	if err != nil {
		return nil, err
	}

	for _, encryptedData := range encryptedDataArray {
		sub, err := Subtract(encryptedData, mean)
		if err != nil {
			return nil, err
		}

		pow, err := Pow2(sub)
		if err != nil {
			return nil, err
		}

		ciphertextPow, err := unmarshallIntoNewCiphertext(pow)
		if err != nil {
			return nil, err
		}

		CkksEvaluator.Relinearize(ciphertextPow, ciphertextPow)

		CkksEvaluator.Add(ciphertextSum, ciphertextPow, ciphertextSum)
	}

	sumSquaredDiff, err := ciphertextSum.MarshalBinary()
	result, err := DivByConst(sumSquaredDiff, float64(len(encryptedDataArray)))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ArithmeticProgressionElementN(firstMember []byte, dif []byte, numberOfMembers []byte) ([]byte, error) {
	dec, err := SubtractConst(numberOfMembers, 1)
	if err != nil {
		return nil, err
	}

	mult, err := MultOf2(dif, dec)
	if err != nil {
		return nil, err
	}

	return SumOf2(firstMember, mult)
}

func ArithmeticProgressionSum(firstMember []byte, dif []byte, numberOfMembers []byte) ([]byte, error) {
	elementN, err := ArithmeticProgressionElementN(firstMember, dif, numberOfMembers)
	if err != nil {
		return nil, err
	}

	sum, err := SumOf2(firstMember, elementN)
	if err != nil {
		return nil, err
	}

	sumCiphertext, err := unmarshallIntoNewCiphertext(sum)
	if err != nil {
		return nil, err
	}
	CkksEvaluator.Relinearize(sumCiphertext, sumCiphertext)

	sum, err = sumCiphertext.MarshalBinary()
	if err != nil {
		return nil, err
	}

	mult, err := MultOf2(numberOfMembers, sum)
	if err != nil {
		return nil, err
	}

	return DivByConst(mult, 2.0)
}

func Covariance(encryptedDataArray1 [][]byte, encryptedDataArray2 [][]byte) ([]byte, error) {
	if len(encryptedDataArray1) != len(encryptedDataArray2) {
		return nil, errors.New("arrays must be of the same length")
	}
	if len(encryptedDataArray1) == 0 || len(encryptedDataArray2) == 0 {
		return nil, errors.New("cannot use empty array")
	}

	mean1, err := ArrayMean(encryptedDataArray1)
	if err != nil {
		return nil, err
	}

	mean2, err := ArrayMean(encryptedDataArray2)
	if err != nil {
		return nil, err
	}

	ciphertextSum, err := MakeZeroCiphertext(encryptedDataArray1[0])
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(encryptedDataArray1); i++ {
		sub1, err := Subtract(encryptedDataArray1[i], mean1)
		if err != nil {
			return nil, err
		}

		sub2, err := Subtract(encryptedDataArray2[i], mean2)
		if err != nil {
			return nil, err
		}

		mult, err := MultOf2(sub1, sub2)
		if err != nil {
			return nil, err
		}

		ciphertextMult, err := unmarshallIntoNewCiphertext(mult)
		if err != nil {
			return nil, err
		}

		CkksEvaluator.Relinearize(ciphertextMult, ciphertextMult)

		CkksEvaluator.Add(ciphertextSum, ciphertextMult, ciphertextSum)
	}

	sum, err := ciphertextSum.MarshalBinary()
	result, err := DivByConst(sum, float64(len(encryptedDataArray1)))
	if err != nil {
		return nil, err
	}

	return result, nil
}

/*
func Inverse(encryptedData []byte, iterations int, initialApproximation float64) ([]byte, error) {
	ciphertext, err := unmarshallIntoNewCiphertext(encryptedData)
	if err != nil {
		return nil, err
	}

	x0 := CkksEvaluator.MultByConstNew(ciphertext, 1.0/initialApproximation)
	if x0.Level() > 0 {
		err := CkksEvaluator.Rescale(ciphertext, CkksParams.DefaultScale(), x0)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < iterations; i++ {
		// x_{n+1} = x_n * (2 - c * x_n)
		cTimesXn := CkksEvaluator.MulRelinNew(ciphertext, x0)

		if cTimesXn.Level() > 0 {
			err := CkksEvaluator.Rescale(cTimesXn, CkksParams.DefaultScale(), cTimesXn)
			if err != nil {
				return nil, err
			}
		}

		twoMinusCTXn := CkksEvaluator.AddConstNew(cTimesXn, -2.0)
		CkksEvaluator.Neg(twoMinusCTXn, twoMinusCTXn)
		xnPlusOne := CkksEvaluator.MulRelinNew(x0, twoMinusCTXn)

		if xnPlusOne.Level() > 0 {
			err := CkksEvaluator.Rescale(xnPlusOne, CkksParams.DefaultScale(), xnPlusOne)
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, err
		}
		x0 = xnPlusOne
	}

	return x0.MarshalBinary()
}
*/
/*
func SqrtOnEncryptedData(ciphertextA *ckks.Ciphertext, iterations int, initialGuess float64, evaluator *ckks.Evaluator, encoder ckks.Encoder, params ckks.Parameters) *ckks.Ciphertext {
	// Начальное приближение для корня
	initialGuessCiphertext := MakeCiphertextFromFloat(initialGuess, evaluator, encoder, params)

	// Итеративное уточнение приближения
	for i := 0; i < iterations; i++ {
		// Вычисление a / x_n
		ciphertextRatio := evaluator.DivNew(ciphertextA, initialGuessCiphertext)

		// Вычисление x_n + (a / x_n)
		ciphertextSum := evaluator.AddNew(initialGuessCiphertext, ciphertextRatio)

		// Вычисление (x_n + (a / x_n)) / 2
		ciphertextSqrt := evaluator.MultByConstNew(ciphertextSum, 0.5)

		// Обновление приближения
		initialGuessCiphertext = ciphertextSqrt
	}

	return initialGuessCiphertext
}

func Divide(encryptedData []byte, encryptedData2 []byte, iterations int, initApr float64, ckksParams ckks.Parameters) ([]byte, error) {
	evaluator := getNewEvaluator(ckksParams)

	ciphertextA := ckks.NewCiphertext(ckksParams, 1, ckksParams.MaxLevel(), ckksParams.DefaultScale())
	ciphertextB := ckks.NewCiphertext(ckksParams, 1, ckksParams.MaxLevel(), ckksParams.DefaultScale())

	err := ciphertextA.UnmarshalBinary(encryptedData)
	if err != nil {
		return nil, err
	}

	err = ciphertextB.UnmarshalBinary(encryptedData2)
	if err != nil {
		return nil, err
	}

	ciphertextInvB := MakeCiphertextFromFloat(initApr, encryptedData, evaluator, ckksParams)

	for i := 0; i < iterations; i++ {
		// tmp = b * x_n  (где x_n - current apr)
		tmp := evaluator.MulNew(ciphertextB, ciphertextInvB)

		//get 2 in cipher text
		twoCiphertext := MakeCiphertextFromFloat(2.0, encryptedData, evaluator, ckksParams)

		// tmp = 2 - tmp  (2 - b * x_n)
		tmp = evaluator.SubNew(twoCiphertext, tmp)

		// x_n+1 = x_n * tmp  (x_n * (2 - b * x_n))
		ciphertextInvB = evaluator.MulNew(ciphertextInvB, tmp)
	}
	ciphertextResult := evaluator.MulNew(ciphertextA, ciphertextInvB)
	return ciphertextResult.MarshalBinary()
}
*/
