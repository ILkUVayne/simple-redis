package src

import "testing"

func TestChangeLoadFactor(t *testing.T) {
	s := new(SRedisServer)
	s.aofChildPid = -1
	s.rdbChildPid = -1
	s.loadFactor = LOAD_FACTOR
	s.changeLoadFactor(LOAD_FACTOR)
	if s.loadFactor != LOAD_FACTOR {
		t.Error("changeLoadFactor err: loadFactor = ", s.loadFactor)
	}
	s.changeLoadFactor(BG_PERSISTENCE_LOAD_FACTOR)
	if s.loadFactor != LOAD_FACTOR {
		t.Error("changeLoadFactor err: loadFactor = ", s.loadFactor)
	}
	s.rdbChildPid = 1
	s.changeLoadFactor(BG_PERSISTENCE_LOAD_FACTOR)
	if s.loadFactor != BG_PERSISTENCE_LOAD_FACTOR {
		t.Error("changeLoadFactor err: loadFactor = ", s.loadFactor)
	}
	s.changeLoadFactor(LOAD_FACTOR)
	if s.loadFactor != BG_PERSISTENCE_LOAD_FACTOR {
		t.Error("changeLoadFactor err: loadFactor = ", s.loadFactor)
	}
}
