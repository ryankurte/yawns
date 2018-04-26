package types

// Frequency type for parsing/rendering
type Frequency float64

func (f *Frequency) MarshalText() ([]byte, error) {
	return MarshalUnit("Hz", float64(*f))
}

func (f *Frequency) UnmarshalText(text []byte) error {
	val, err := UnmarshalUnit("Hz", text)
	if err != nil {
		return err
	}
	*f = Frequency(val)
	return nil
}

// Distance type for parsing/rendering
type Distance float64

func (d *Distance) MarshalText() ([]byte, error) {
	return MarshalUnit("m", float64(*d))
}

func (d *Distance) UnmarshalText(text []byte) error {
	val, err := UnmarshalUnit("m", text)
	if err != nil {
		return err
	}
	*d = Distance(val)
	return nil
}

// Attenuation type for parsing/rendering
type Attenuation float64

func (a *Attenuation) MarshalText() ([]byte, error) {
	return MarshalUnit("dB", float64(*a))
}

func (a *Attenuation) UnmarshalText(text []byte) error {
	val, err := UnmarshalUnit("dB", text)
	if err != nil {
		return err
	}
	*a = Attenuation(val)
	return nil
}

// AttenuationMap is a map of attenuation values with keys
type AttenuationMap map[string]Attenuation

func (am AttenuationMap) Reduce() Attenuation {
	sum := Attenuation(0)
	for _, v := range am {
		sum += v
	}
	return sum
}

// Baud type for parsing/rendering
type Baud float64

func (b *Baud) MarshalText() ([]byte, error) {
	return MarshalUnit("bps", float64(*b))
}

func (b *Baud) UnmarshalText(text []byte) error {
	val, err := UnmarshalUnit("bps", text)
	if err != nil {
		return err
	}
	*b = Baud(val)
	return nil
}

func (b *Baud) String() string {
	val, _ := MarshalUnit("B", float64(*b))
	return string(val)
}

type SizeBytes uint32

func (b *SizeBytes) MarshalText() ([]byte, error) {
	return MarshalUnit("B", float64(*b))
}

func (b *SizeBytes) UnmarshalText(text []byte) error {
	val, err := UnmarshalUnit("B", text)
	if err != nil {
		return err
	}
	*b = SizeBytes(val)
	return nil
}

func (b *SizeBytes) String() string {
	val, _ := MarshalUnit("B", float64(*b))
	return string(val)
}
