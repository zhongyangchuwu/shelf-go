package vault

type Level string

const (
	LevelOK   Level = "ok"
	LevelWarn Level = "warn"
	LevelFail Level = "fail"
)

type Check struct {
	Level  Level
	Name   string
	Detail string
}

type Report []Check

func (r *Report) OK(name, detail string) {
	*r = append(*r, Check{Level: LevelOK, Name: name, Detail: detail})
}

func (r *Report) Warn(name, detail string) {
	*r = append(*r, Check{Level: LevelWarn, Name: name, Detail: detail})
}

func (r *Report) Fail(name, detail string) {
	*r = append(*r, Check{Level: LevelFail, Name: name, Detail: detail})
}

func HasFailures(report Report) bool {
	for _, check := range report {
		if check.Level == LevelFail {
			return true
		}
	}
	return false
}
