package logger

import "github.com/sirupsen/logrus"

func New(loglevelRaw string) (*logrus.Logger, error) {
	lvl, err := logrus.ParseLevel(loglevelRaw)
	if err != nil {
		panic(err)
	}

	log := logrus.New()
	log.SetLevel(lvl)
	return log, nil
}
