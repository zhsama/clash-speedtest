package unlock

import (
	"sort"
	"sync"
)

var (
	detectors = make(map[string]UnlockDetector)
	mu        sync.RWMutex
)

// Register registers a detector for automatic loading
func Register(detector UnlockDetector) {
	mu.Lock()
	defer mu.Unlock()
	if detector == nil {
		panic("cannot register a nil detector")
	}
	name := detector.GetPlatformName()
	if _, exists := detectors[name]; exists {
		panic("detector already registered: " + name)
	}
	detectors[name] = detector
}

// GetDetectors returns all registered detectors
func GetDetectors() map[string]UnlockDetector {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]UnlockDetector, len(detectors))
	for name, detector := range detectors {
		result[name] = detector
	}
	return result
}

// GetDetectorsByPriority returns detectors sorted by priority
func GetDetectorsByPriority() []UnlockDetector {
	mu.RLock()
	defer mu.RUnlock()
	
	detectorList := make([]UnlockDetector, 0, len(detectors))
	for _, detector := range detectors {
		detectorList = append(detectorList, detector)
	}
	
	sort.Slice(detectorList, func(i, j int) bool {
		return detectorList[i].GetPriority() < detectorList[j].GetPriority()
	})
	
	return detectorList
}

// GetDetector returns a specific detector by name
func GetDetector(name string) (UnlockDetector, bool) {
	mu.RLock()
	defer mu.RUnlock()
	detector, exists := detectors[name]
	return detector, exists
}

// GetRegisteredPlatforms returns all registered platform names
func GetRegisteredPlatforms() []string {
	mu.RLock()
	defer mu.RUnlock()
	platforms := make([]string, 0, len(detectors))
	for name := range detectors {
		platforms = append(platforms, name)
	}
	return platforms
}