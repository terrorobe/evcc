# Dynamic Home Battery SoC Optimization Feature

## Overview

This feature introduces intelligent **dynamic State of Charge (SoC) optimization** for home batteries that automatically adapts to electricity price patterns and solar generation forecasts. Rather than fixed-time schedules, the system dynamically determines optimal SoC targets to ride through price peaks while minimizing unnecessary grid charging before sunny periods.

## Background & Motivation

### The Dynamic Duck Curve Challenge

Electricity markets exhibit seasonal "duck curve" variations with distinct patterns:

#### Winter Pattern
- **Noon vs Night**: Similar low prices (both off-peak)
- **Morning/Evening**: High price bumps during demand peaks
- **Solar**: Limited generation, less price impact

#### Summer Pattern
- **Noon**: Near-zero or negative prices (high solar)
- **Night**: Moderately low prices (higher than summer noon)
- **Morning/Evening**: High price bumps persist
- **Solar**: High generation potential during low-price periods

#### Key Insights
1. **Peak Avoidance**: Morning/evening bumps are consistent across seasons
2. **Seasonal Charging**: Night vs noon charging preference varies by season
3. **Solar Synergy**: Avoid grid charging before high solar generation days
4. **Dynamic Targets**: Optimal SoC varies based on upcoming price patterns and solar forecast

### Current State Analysis
EVCC already provides:
- ✅ Sophisticated car charging plans with cost optimization
- ✅ Home battery monitoring and mode control
- ✅ Comprehensive tariff integration (20+ providers)
- ✅ Solar forecasting capabilities (Solcast integration)
- ✅ Real-time price pattern analysis
- ❌ **Missing**: Dynamic battery SoC optimization based on price patterns
- ❌ **Missing**: Solar-aware battery charging strategies

## Feature Requirements

### Functional Requirements

#### FR1: Dynamic Price Pattern Analysis
- **FR1.1**: Automatically detect morning and evening price peaks
- **FR1.2**: Identify low-price periods (noon vs night) with seasonal adaptation
- **FR1.3**: Calculate peak duration and severity for SoC planning
- **FR1.4**: Adapt to different tariff providers and price patterns

#### FR2: Solar-Aware SoC Optimization
- **FR2.1**: Integrate solar generation forecasts into charging decisions
- **FR2.2**: Minimize grid charging before high solar generation periods
- **FR2.3**: Balance grid charging vs solar charging opportunities
- **FR2.4**: Adjust SoC targets based on multi-day solar forecasts
- **FR2.5**: Handle forecast uncertainty with conservative strategies

#### FR3: Dynamic SoC Target Calculation
- **FR3.1**: Calculate minimum SoC to ride through identified price peaks
- **FR3.2**: Optimize charging timing based on price and solar patterns
- **FR3.3**: Avoid excessive grid charging before sunny periods
- **FR3.4**: Dynamic adjustment based on consumption patterns
- **FR3.5**: Emergency charging when targets cannot be met optimally

#### FR4: Integration with Existing Systems
- **FR4.1**: Reuse existing tariff providers and rate data
- **FR4.2**: Leverage existing site-level coordination with EV charging
- **FR4.3**: Respect existing battery priority/buffer SoC settings
- **FR4.4**: Leverage existing solar forecast infrastructure

#### FR5: User Interface and Control
- **FR5.1**: Single economic threshold configuration (minSavingsPerKwh)
- **FR5.2**: Visual display of price patterns and charging decisions
- **FR5.3**: Real-time optimization status and reasoning
- **FR5.4**: Simple enable/disable override controls
- **FR5.5**: Historical performance and cost savings analytics

### Non-Functional Requirements

#### NFR1: Performance
- **NFR1.1**: Plan calculation within 5 seconds
- **NFR1.2**: Minimal impact on existing system performance
- **NFR1.3**: Efficient rate data reuse from existing car plans

#### NFR2: Reliability
- **NFR2.1**: Graceful degradation when tariff data unavailable
- **NFR2.2**: Automatic plan adjustment for unexpected consumption
- **NFR2.3**: Fallback to simple time-based charging if optimization fails

#### NFR3: Usability
- **NFR3.1**: Configuration complexity similar to existing car plans
- **NFR3.2**: Clear visual feedback on plan effectiveness
- **NFR3.3**: Intuitive default values and smart suggestions

## Technical Design

### Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Tariff Data   │    │ Price Pattern    │    │ Dynamic SoC     │
│  (Existing)     │──→ │   Analyzer       │──→ │  Optimizer      │
└─────────────────┘    │     (New)        │    │    (New)        │
                       └──────────────────┘    └─────────────────┘
                                │                        │
                                │                        ▼
                                │               ┌─────────────────┐
                                │               │ Battery Mode    │
                                │               │ Decision        │
                                │               │    (New)        │
                                │               └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │  Solar Forecast  │    │ Existing Battery│
                       │   (Existing)     │    │ Control System  │
                       └──────────────────┘    │   (Existing)    │
                                │              └─────────────────┘
                                │                        │
                                └────────────────────────┘
                                    Current SoC Reading
```

### Core Components

#### 1. Price Pattern Analyzer
```go
// New component: core/analyzer/price_pattern.go
type PricePatternAnalyzer struct {
    log   *util.Logger
    rates api.Rates
}

type PricePeak struct {
    Start     time.Time     `json:"start"`
    End       time.Time     `json:"end"`
    Duration  time.Duration `json:"duration"`  // Peak duration
    AvgPrice  float64       `json:"avgPrice"`
    Severity  float64       `json:"severity"`  // Price ratio vs daily average
    Type      string        `json:"type"`      // "morning" or "evening"
}

type PricePattern struct {
    Date          time.Time   `json:"date"`
    DailyAverage  float64     `json:"dailyAverage"`
    NoonAverage   float64     `json:"noonAverage"`    // 10:00-14:00
    NightAverage  float64     `json:"nightAverage"`   // 00:00-06:00
    Peaks         []PricePeak `json:"peaks"`
    Season        string      `json:"season"`         // "summer" or "winter"
}

func (ppa *PricePatternAnalyzer) AnalyzePattern(rates api.Rates) PricePattern {
    pattern := PricePattern{
        Date: time.Now(),
    }

    // Calculate adaptive thresholds based on actual price data
    pattern.DailyAverage = ppa.calculateAverage(rates)
    pattern.NoonAverage = ppa.calculateTimeAverage(rates, 10, 14) // 10:00-14:00
    pattern.NightAverage = ppa.calculateTimeAverage(rates, 0, 6)  // 00:00-06:00

    // Detect peaks using configurable ratio relative to daily average
    pattern.Peaks = ppa.DetectPeaks(rates)

    // Determine season based on noon/night price relationship
    pattern.Season = ppa.ClassifySeason(pattern)

    return pattern
}

func (ppa *PricePatternAnalyzer) DetectPeaks(rates api.Rates) []PricePeak {
    // Calculate adaptive thresholds based on actual price data
    dailyAverage := ppa.calculateAverage(rates)
    peakThreshold := dailyAverage * 1.4 // Configurable ratio
    return ppa.findPeaksAboveThreshold(rates, peakThreshold)
}

func (ppa *PricePatternAnalyzer) ClassifySeason(pattern PricePattern) string {
    // Determine season based on noon/night price relationship
    noonToNightRatio := pattern.NoonAverage / pattern.NightAverage
    
    // These thresholds are adaptive - they could be learned over time
    if noonToNightRatio < 0.7 {
        return "summer" // Noon significantly cheaper (high solar impact)
    }
    if noonToNightRatio > 0.9 {
        return "winter" // Similar prices (low solar impact)
    }
    return "transitional" // Spring/fall
}
```

#### 1.1. Consumption Estimator
```go
// Conservative consumption forecasting without complex weather modeling
type ConsumptionEstimator struct {
    historicalData []float64  // kWh per day for last 30 days
    config         ConsumptionConfig
}

func (ce *ConsumptionEstimator) estimateDailyConsumption() float64 {
    if len(ce.historicalData) < 7 {
        // Not enough history, use static fallback
        consumption := ce.config.StaticFallbackKwh
        if ce.isWeekend() {
            consumption *= ce.config.WeekendReduction
        }
        return consumption
    }

    // Use 90th percentile for conservative estimate
    return ce.calculatePercentile(ce.historicalData, 0.9)
}
```

#### 2. Configuration Overview

The dynamic battery optimization is configured through the site configuration YAML (detailed schema provided in Configuration Schema section).

#### 3. Dynamic SoC Optimizer
```go
// Located in core/optimizer/battery_soc.go
type DynamicSocOptimizer struct {
    log                 *util.Logger
    config              DynamicBatteryConfig
    patternAnalyzer     *PricePatternAnalyzer
    consumptionEstimator *ConsumptionEstimator
    socRangeManager     *SocRangeManager
    solarForecast       api.Rates  // Solar generation forecast
    battery             api.Battery
    site                *Site      // To set battery mode
    capacity            float64
}

type ModeDecision struct {
    BatteryMode   api.BatteryMode `json:"batteryMode"`   // Required battery mode
    TargetSoc     float64         `json:"targetSoc"`     // Target SoC for decision
    Reasoning     string          `json:"reasoning"`     // User-facing explanation of why this mode was chosen
    Strategy      string          `json:"strategy"`      // "peak-riding", "solar-aware", "cost-optimal"
    ValidUntil    time.Time       `json:"validUntil"`    // When to recalculate
}

func (dso *DynamicSocOptimizer) OptimizeBatteryMode(
    currentSoc float64,
    pricePattern PricePattern,
    solarForecast api.Rates,
    consumptionForecast float64,
) ModeDecision {
    targetSoc := dso.calculateOptimalTargetSoc(pricePattern, currentSoc)
    batteryMode := dso.calculateRequiredBatteryMode(currentSoc, targetSoc, pricePattern)

    return ModeDecision{
        BatteryMode: batteryMode,
        TargetSoc:   targetSoc,
        Reasoning:   dso.explainDecision(pricePattern, targetSoc, batteryMode),
        Strategy:    dso.determineStrategy(pricePattern),
        ValidUntil:  dso.calculateNextRecalculation(),
    }
}
```

### Algorithm Details

#### Mode-Based Optimization Algorithm

The core optimization algorithm calculates optimal battery modes based on price patterns and current SoC conditions:

##### 1. Core Battery Mode Decision Strategy

```go
func (dso *DynamicSocOptimizer) calculateRequiredBatteryMode(
    currentSoc float64,
    targetSoc float64,
    pattern PricePattern,
) api.BatteryMode {
    currentTime := time.Now()

    // During cheap periods: encourage charging to target
    if dso.isInCheapPeriod(pattern, currentTime) && currentSoc < targetSoc {
        return api.BatteryCharge  // Force charging to reach target
    }

    // During expensive periods: allow discharge
    if dso.isInExpensivePeriod(pattern, currentTime) {
        return api.BatteryNormal  // Allow natural discharge
    }

    // Strategic decisions for upcoming peaks
    upcomingPeak := dso.getNextExpensivePeriod(pattern, currentTime)
    if upcomingPeak != nil {
        currentPrice := dso.getCurrentPrice(pattern, currentTime)
        
        // Strategic charging during moderate prices before expensive peaks
        if currentSoc < targetSoc && dso.isStrategicChargingBeneficial(currentPrice, upcomingPeak.AvgPrice) {
            return api.BatteryCharge
        }
        
        // Strategic holding: preserve SoC when beneficial for upcoming peaks
        requiredPeakRidingSoc := dso.calculatePeakRidingSoc(pattern.Peaks)
        if currentPrice < upcomingPeak.AvgPrice && currentSoc <= requiredPeakRidingSoc {
            return api.BatteryHold
        }
    }

    // At target SoC: hold to preserve energy for peaks
    if math.Abs(currentSoc - targetSoc) < 2.0 {
        return api.BatteryHold    // Maintain current SoC
    }

    // Default: normal operation
    return api.BatteryNormal
}

func (dso *DynamicSocOptimizer) calculateOptimalTargetSoc(
    pattern PricePattern,
    currentSoc float64,
) float64 {
    // Calculate SoC needed to ride through upcoming peaks
    peakRidingTarget := dso.calculatePeakRidingSoc(pattern.Peaks)

    // Apply solar-aware adjustments
    if dso.shouldAvoidChargingBeforeSolar() {
        peakRidingTarget = math.Min(peakRidingTarget, float64(dso.config.SocRange.PreferredMax))
    }

    // Apply SoC range constraints and return
    return dso.socRangeManager.applySocRangeConstraints(peakRidingTarget, currentSoc)
}
```

##### 2. Peak Riding SoC Calculation
**Objective**: Calculate sufficient SoC to ride through identified price peaks

```go
func (dso *DynamicSocOptimizer) calculatePeakRidingSoc(peaks []PricePeak) float64 {
    if len(peaks) == 0 {
        return float64(dso.config.SocRange.PreferredMin) // No peaks, use minimum
    }

    // Find the longest/most severe peak to prepare for
    criticalPeak := dso.findMostCriticalPeak(peaks)

    // Calculate energy needed to ride through this peak
    estimatedConsumption := dso.getEstimatedConsumption() // kWh per hour
    energyRequired := estimatedConsumption * criticalPeak.Duration.Hours()

    // Account for round-trip efficiency losses
    energyToStore := energyRequired / dso.config.Battery.RoundTripEfficiency

    // Convert to SoC percentage with safety margin
    targetSoc := (energyToStore / dso.capacity) * 100
    safetyMargin := float64(dso.config.Optimization.SafetyMarginPercent)

    return targetSoc + safetyMargin
}

func (dso *DynamicSocOptimizer) shouldAvoidChargingBeforeSolar() bool {
    tomorrowSolar := dso.getTomorrowSolarGeneration()
    forecastConfidence := dso.getSolarForecastConfidence()

    // Only use solar strategy if forecast is reliable enough
    if forecastConfidence < dso.config.SolarAware.ForecastConfidenceMin {
        return false
    }

    // Check if tomorrow's solar is significant relative to battery capacity
    significanceThreshold := dso.capacity * dso.config.SolarAware.SignificanceRatio
    return tomorrowSolar > significanceThreshold
}
```

##### 3. Economic Viability Check
**Objective**: Only set high minSoC targets when economically beneficial

```go
func (dso *DynamicSocOptimizer) isInCheapPeriod(
    pattern PricePattern,
    currentTime time.Time,
) bool {
    currentPrice := dso.getCurrentPrice(pattern, currentTime)

    // Check if current price is economically viable for charging
    effectiveCost := dso.calculateEffectiveCost(currentPrice)

    upcomingPeaks := dso.getUpcomingPeaks(pattern.Peaks)
    if len(upcomingPeaks) == 0 {
        return false // No upcoming peaks to prepare for
    }

    avgPeakPrice := dso.calculateAveragePeakPrice(upcomingPeaks)
    savings := avgPeakPrice - effectiveCost

    if savings <= dso.config.Optimization.MinSavingsPerKwh {
        return false
    }

    // Additional check: current price should be below daily average
    return currentPrice < pattern.DailyAverage
}

func (dso *DynamicSocOptimizer) isInExpensivePeriod(
    pattern PricePattern,
    currentTime time.Time,
) bool {
    currentPrice := dso.getCurrentPrice(pattern, currentTime)

    // Consider expensive if price is above daily average + margin
    expensiveThreshold := pattern.DailyAverage * 1.2 // 20% above average
    return currentPrice > expensiveThreshold
}

func (dso *DynamicSocOptimizer) getNextExpensivePeriod(
    pattern PricePattern,
    currentTime time.Time,
) *PricePeak {
    for _, peak := range pattern.Peaks {
        if peak.Start.After(currentTime) {
            return &peak
        }
    }
    return nil
}

func (dso *DynamicSocOptimizer) calculateEffectiveCost(price float64) float64 {
    return (price / dso.config.Battery.RoundTripEfficiency) + dso.config.Battery.WearCostPerKwh
}

func (dso *DynamicSocOptimizer) isStrategicChargingBeneficial(
    currentPrice float64,
    peakPrice float64,
) bool {
    effectiveCost := dso.calculateEffectiveCost(currentPrice)
    savings := peakPrice - effectiveCost
    return savings > dso.config.Optimization.MinSavingsPerKwh
}
```

#### Round-Trip Efficiency Integration

**Key Principle**: All optimization decisions must account for the fact that storing and retrieving energy from the battery is lossy.

**Impact on Strategy Selection**:
- **Negative Prices**: Always beneficial regardless of efficiency
- **Small Price Differences**: May not be worth it after efficiency losses
- **Peak Avoidance**: Higher SoC targets needed to compensate for losses

#### SoC Range Management for Battery Health

**Key Principle**: Balance economic optimization with battery longevity by maintaining preferred SoC ranges.

```go
type SocRangeManager struct {
    config             DynamicBatteryConfig
    timeOutsideRange   time.Duration
    lastRangeCheckTime time.Time
}

func (srm *SocRangeManager) applySocRangeConstraints(
    targetSoc float64,
    currentSoc float64,
) float64 {
    // Never exceed absolute emergency limits
    if targetSoc > float64(srm.config.SocRange.EmergencyMax) {
        return float64(srm.config.SocRange.EmergencyMax)
    }
    if targetSoc < float64(srm.config.SocRange.EmergencyMin) {
        return float64(srm.config.SocRange.EmergencyMin)
    }

    // Check if we've been outside preferred range too long
    if srm.hasBeenOutsideRangeTooLong(currentSoc) {
        return srm.forceReturnToPreferredRange(currentSoc, targetSoc)
    }

    // Normal case: allow temporary excursions outside preferred range
    return targetSoc
}

func (srm *SocRangeManager) hasBeenOutsideRangeTooLong(currentSoc float64) bool {
    preferredMin := float64(srm.config.SocRange.PreferredMin)
    preferredMax := float64(srm.config.SocRange.PreferredMax)

    // Update time tracking
    now := time.Now()
    if currentSoc < preferredMin || currentSoc > preferredMax {
        if srm.isFirstTimeOutsideRange() {
            srm.lastRangeCheckTime = now // Start tracking
        }
        srm.timeOutsideRange = now.Sub(srm.lastRangeCheckTime)
    } else {
        srm.timeOutsideRange = 0 // Reset when back in range
    }

    maxDuration := time.Duration(srm.config.SocRange.MaxDurationOutsideRange) * time.Hour
    return srm.timeOutsideRange > maxDuration
}

func (srm *SocRangeManager) forceReturnToPreferredRange(
    currentSoc float64,
    targetSoc float64,
) float64 {
    preferredMin := float64(srm.config.SocRange.PreferredMin)
    preferredMax := float64(srm.config.SocRange.PreferredMax)

    // Force return to preferred range
    if currentSoc < preferredMin {
        // Been too low too long - force charge to preferred minimum
        return math.Max(targetSoc, preferredMin)
    }
    if currentSoc > preferredMax {
        // Been too high too long - force discharge by setting low target
        return math.Min(targetSoc, preferredMax)
    }

    return targetSoc // Already in range
}
```

**SoC Range Management Examples**:

1. **Normal Operation** (45% SoC, 30-80% range): ✅ Optimization proceeds normally
2. **Temporary Excursion** (85% SoC, 3h outside range): ✅ Allow natural discharge
3. **Force Return** (85% SoC, 14h outside range): ⚠️ Force discharge regardless of economics
4. **Emergency Protection** (Target 25%, Emergency min 10%): ⚠️ Clamp to preferred minimum (30%)

##### Optimization Scheduling

```go
// Optimization recalculation frequency management
type OptimizationScheduler struct {
    recalculateInterval time.Duration // 15 minutes for normal operation
    priceUpdateTrigger  bool          // Immediate recalc when new prices arrive at 14:00
    socChangeThreshold  float64       // Recalc if SoC changes >5%
    lastBatteryMode     api.BatteryMode // Track when mode actually changes
    lastSocReading      float64       // Track SoC changes for recalculation triggers
}

func (os *OptimizationScheduler) shouldRecalculate(
    currentSoc float64,
    lastCalculation time.Time,
) bool {
    // Always recalculate when new price data arrives (daily at 14:00)
    if os.priceUpdateTrigger {
        return true
    }

    // Recalculate on significant SoC changes
    socChange := math.Abs(currentSoc - os.lastSocReading)
    if socChange > os.socChangeThreshold {
        return true
    }

    // Regular interval recalculation
    return time.Since(lastCalculation) > os.recalculateInterval
}
```

#### 4. Integration Points

##### Site-Level Integration
**Integration with `core/site_battery.go`**

The dynamic battery optimizer integrates directly into the existing `requiredBatteryMode` function:

```go
// Enhanced requiredBatteryMode in core/site_battery.go
func (site *Site) requiredBatteryMode(batteryGridChargeActive bool, rate api.Rate) api.BatteryMode {
    // ... existing logic for external mode, reset logic ...

    switch {
    case !site.batteryConfigured():
        res = api.BatteryUnknown
    case extModeReset:
        res = api.BatteryNormal
    case extMode != api.BatteryUnknown:
        if extMode != batMode {
            res = extMode
        }
    case site.dynamicBatteryOptimizationActive():
        // NEW: Dynamic optimization takes priority over legacy grid charge logic
        res = site.requiredBatteryModeOptimized(rate)
    case batteryGridChargeActive:
        res = mapper(api.BatteryCharge)
    case site.dischargeControlActive(rate):
        res = mapper(api.BatteryHold)
    case batteryModeModified(batMode):
        res = api.BatteryNormal
    }

    return res
}

// NEW: Dynamic battery mode determination
func (site *Site) requiredBatteryModeOptimized(rate api.Rate) api.BatteryMode {
    // Emergency charging always takes priority
    emergencyThreshold := float64(site.config.DynamicBatteryOptimization.SocRange.EmergencyMin)
    if site.batterySoc <= emergencyThreshold {
        return api.BatteryCharge
    }

    // Get optimization decision
    pricePattern := site.dynamicBatteryOptimizer.patternAnalyzer.AnalyzePattern(site.rates)
    decision := site.dynamicBatteryOptimizer.OptimizeBatteryMode(
        site.batterySoc, pricePattern, site.solarForecast, site.estimatedConsumption,
    )

    // Store the target SoC for UI display and status reporting
    site.dynamicBatteryTargetSoc = decision.TargetSoc
    site.dynamicBatteryDecision = decision

    return decision.BatteryMode
}

// Check if dynamic optimization should be active
func (site *Site) dynamicBatteryOptimizationActive() bool {
    return site.config.DynamicBatteryOptimization.Enable && 
           site.dynamicBatteryOptimizer != nil &&
           site.rates != nil // Require rate data for optimization
}
```

**Integration Benefits:**
- ✅ **Preserves existing behavior** when optimization is disabled
- ✅ **Respects external battery mode** - external control takes priority  
- ✅ **Maintains emergency safety** - emergency charging logic preserved
- ✅ **Inherits coordination** - EV charging, discharge control handled by existing logic
- ✅ **Single entry point** - all battery decisions flow through `requiredBatteryMode`

##### Emergency Override Integration
```go
// Extends core/site_battery.go
func (site *Site) checkEmergencyCharging() {
    emergencyThreshold := float64(site.config.DynamicBatteryOptimization.SocRange.EmergencyMin)

    if site.batterySoc <= emergencyThreshold {
        site.log.WARN.Printf("Emergency charging activated: SoC %.1f%% <= %.1f%% threshold",
                           site.batterySoc, emergencyThreshold)

        // Override any optimization - force emergency charging
        site.SetBatteryMode(api.BatteryCharge)
        site.currentBatteryMode = api.BatteryCharge
    }
}
```


### Battery Mode Decision Examples

#### Example 1: Normal Day with Evening Peak
```
Time: 10:00
Current SoC: 45%
Price Pattern: Cheap until 14:00 (€0.15), expensive 18:00-21:00 (€0.45)
Solar Forecast: Low (3kWh)

Algorithm Decision:
- Currently in cheap period: ✅
- Economic viability: (€0.15 ÷ 0.85) + €0.10 = €0.28 effective cost
- Peak price: €0.45, savings: €0.17 > €0.02 minimum ✅
- Target SoC: 75% needed for evening peak
- Current SoC: 45% < 75% target

Battery Mode: api.BatteryCharge
Reasoning: "Cheap period - charging to 75% for evening peak (saves €0.17/kWh)"
```

#### Example 2: Expensive Period - Allow Discharge
```
Time: 19:00
Current SoC: 80%
Price Pattern: Currently in evening peak (€0.45)
Consumption: Using battery to avoid expensive grid power

Algorithm Decision:
- Currently in expensive period: ✅
- Allow discharge to avoid expensive grid power

Battery Mode: api.BatteryNormal
Reasoning: "Expensive period - allowing discharge to avoid €0.45/kWh grid power"
```

#### Example 3: Strategic Night Charging
```
Time: 02:00
Current SoC: 50%
Night Price: €0.20/kWh (moderate - not "cheap" but below morning peak)
Morning Peak: 07:00-09:00 (€0.40/kWh)
Target SoC: 75%

Algorithm Decision:
- Not in cheap period: €0.20 > daily average ❌
- Not in expensive period: €0.20 < €0.30 (120% of €0.25 avg) ❌
- Strategic charging check: upcoming peak at 07:00 ✅
- Economic viability: (€0.20 ÷ 0.85) + €0.10 = €0.34 effective cost
- Peak price: €0.40, savings: €0.06 > €0.02 minimum ✅
- Current SoC: 50% < 75% target ✅

Battery Mode: api.BatteryCharge
Reasoning: "Strategic charging before morning peak - moderate price saves €0.06/kWh vs peak"
```

#### Example 4: Strategic Morning Holding
```
Time: 05:00
Current SoC: 40%
Required Peak-riding SoC: 45%
Current Price: €0.28/kWh (expensive - not viable for charging)
Morning Peak: 07:00-09:00 (€0.45/kWh)

Algorithm Decision:
- Not in cheap period: €0.28 > daily average ❌
- In expensive period: €0.28 > €0.30 threshold ✅ 
- But strategic charging not beneficial: (€0.28 ÷ 0.85) + €0.10 = €0.43 effective cost
- Peak price: €0.45, savings: €0.02 = minimum threshold (marginal) ❌
- Strategic holding check: currentPrice (€0.28) < peakPrice (€0.45) ✅
- Current SoC: 40% <= 45% required ✅

Battery Mode: api.BatteryHold
Reasoning: "Preserving SoC for morning peak - current €0.28 < peak €0.45/kWh"
```

#### Example 5: Solar-Aware Charging Avoidance
```
Time: 02:00
Current SoC: 40%
Night Price: €0.18/kWh
Solar Forecast: 22kWh tomorrow (high confidence)
Battery Capacity: 15kWh

Algorithm Decision:
- Economic viability: (€0.18 ÷ 0.85) + €0.10 = €0.31 effective cost
- Solar significance: 22kWh > (15kWh × 1.5) = Yes, significant
- Current SoC above minimum safety: 40% > 30% ✅
- Target SoC: 40% (maintain current)
- Current SoC: 40% ≈ 40% target

Battery Mode: api.BatteryHold
Reasoning: "Avoiding night charging - significant solar expected tomorrow (22kWh)"
```

### Configuration Schema

#### YAML Configuration
```yaml
site:
  # Dynamic Battery SoC Optimization
  dynamicBatteryOptimization:
    enable: true

    # Battery efficiency and physical constraints
    battery:
      roundTripEfficiency: 0.85    # 85% efficiency (15% loss) - configurable
      wearCostPerKwh: 0.10         # €0.10 wear cost per kWh cycled

    # SoC operating range for battery health
    socRange:
      preferredMin: 30             # Preferred minimum SoC (%) - avoid staying below
      preferredMax: 80             # Preferred maximum SoC (%) - avoid staying above
      maxDurationOutsideRange: 12  # Max hours outside preferred range before forcing return
      emergencyMin: 10             # Absolute minimum SoC (never go below)
      emergencyMax: 95             # Absolute maximum SoC (never go above)

    # Optimization behavior
    optimization:
      minSavingsPerKwh: 0.02        # €0.02 minimum savings required to charge
      safetyMarginPercent: 10       # Additional SoC margin above calculated minimum (%)

    # Consumption estimation for SoC target calculation
    consumption:
      method: "conservative"       # Use 90th percentile of historical consumption
      staticFallbackKwh: 20       # Default daily consumption if no history available
      weekendReduction: 0.7       # 30% less consumption on weekends

    # Solar-aware charging (adaptive thresholds)
    solarAware:
      enable: true
      significanceRatio: 1.5       # Solar must be 1.5x+ vs battery capacity to be "significant"
      forecastConfidenceMin: 0.6   # Minimum forecast confidence to use solar strategy
      maxWaitHours: 24             # Max hours to wait for solar before emergency charging

    # Pattern detection (adaptive thresholds)
    patternDetection:
      peakSeverityRatio: 1.4       # Price must be 1.4x+ daily average to be a "peak"
      minPriceSpreadRatio: 0.2     # Min spread (cheapest/most expensive) as ratio of avg price

    # Battery charging constraints (for planning calculations)
    charging:
      maxChargeRateKw: 5.0         # Maximum charging rate in kW (used for time estimates)

    # Note: Emergency charging handled by socRange.emergencyMin threshold
    # When disabled or rate data unavailable, system defers to existing battery logic

```

### API Endpoints

#### Configuration API
```http
GET /api/config/site/dynamicBatteryOptimization
PUT /api/config/site/dynamicBatteryOptimization
```

#### Status API
```http
GET /api/site/batteryOptimization/status
GET /api/site/batteryOptimization/pricePattern
GET /api/site/batteryOptimization/decision
GET /api/site/batteryOptimization/targetSoc
```

**Status Response:**
```json
{
  "enabled": true,
  "currentDecision": {
    "batteryMode": "charge",
    "targetSoc": 75.0,
    "currentSoc": 45.0,
    "strategy": "peak-riding",
    "reasoning": "Cheap period - charging to 75% for evening peak (saves €0.17/kWh)",
    "validUntil": "2024-06-15T14:00:00Z",
    "lastUpdated": "2024-06-15T10:15:00Z"
  },
  "pricePattern": {
    "season": "summer",
    "dailyAverage": 0.25,
    "noonAverage": 0.08,
    "nightAverage": 0.18,
    "peaks": [
      {
        "type": "morning",
        "start": "2024-06-15T07:00:00Z",
        "end": "2024-06-15T09:00:00Z",
        "avgPrice": 0.35,
        "severity": 1.4
      },
      {
        "type": "evening",
        "start": "2024-06-15T18:00:00Z",
        "end": "2024-06-15T21:00:00Z",
        "avgPrice": 0.42,
        "severity": 1.68
      }
    ]
  },
  "solarForecast": {
    "today": 22.5,
    "tomorrow": 18.0,
    "confidence": "high"
  },
  "performance": {
    "costSavingsToday": 3.20,
    "costSavingsWeek": 18.75,
    "peaksAvoided": 14,
    "solarOptimization": 0.85,
    "currency": "EUR"
  }
}
```

## User Interface Design

### Configuration UI

#### Dynamic Battery Optimization Settings
```
┌─────────────────────────────────────────────────────────────┐
│ Dynamic Battery Optimization                        [⚙️]   │
├─────────────────────────────────────────────────────────────┤
│ ☑️ Enable Dynamic SoC Optimization                         │
│                                                             │
│ Cost Optimization                                           │
│ Min savings required: [0.02] €/kWh                         │
│                                                             │
│ Solar Integration                                           │
│ ☑️ Avoid charging before sunny days                        │
│ Solar significance: [1.5]x battery capacity                │
│                                                             │
│ Safety                                                      │
│ Safety margin: [10]%                                       │
│                                                             │
│                                    [Reset] [Save]          │
└─────────────────────────────────────────────────────────────┘
```

### Status Display

#### Dynamic Battery Optimization Dashboard
```
┌─────────────────────────────────────────────────────────────┐
│ Battery Optimization Status              🔋 45% → 🎯 75% │
├─────────────────────────────────────────────────────────────┤
│ Current Strategy: Solar-Aware Charging                     │
│ 🌞 Waiting for tomorrow's 18kWh solar forecast             │
│                                                             │
│ Next Charging: 11:00-13:00 (€0.12/kWh) ⬅️ Mixed solar/grid │
│ Target SoC: 75% for evening peak (18:00-21:00)             │
│                                                             │
│ SoC Health: ✅ In preferred range (30-80%)                 │
│ ██████████████████████▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ 30% ←→ 80%      │
│                                                             │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ Today's Price Pattern (Summer)               [📊 Details]  │
│                                                             │
│ 00:00 ████▓▓▓▓ €0.18  Night                                │
│ 06:00 ████████ €0.35  Morning Peak 🔴                      │
│ 12:00 ▓▓▓▓▓▓▓▓ €0.08  Solar Low 🟢                         │
│ 18:00 ██████▓▓ €0.42  Evening Peak 🔴                      │
│                                                             │
│ 📈 Performance: €3.20 saved today • 14 peaks avoided       │
└─────────────────────────────────────────────────────────────┘
```



## Implementation Plan

### Phase 1: Core Infrastructure
1. **Price Pattern Analyzer**
   - Create `core/analyzer/price_pattern.go`
   - Implement market-adaptive pattern detection
   - Add confidence scoring and fallback logic

2. **Dynamic SoC Optimizer**
   - Create `core/optimizer/battery_soc.go`
   - Implement efficiency-aware cost calculations
   - Add SoC range management for battery health

3. **Configuration Integration**
   - Extend site configuration schema for simplified settings
   - Add YAML parsing with smart defaults
   - Create configuration validation

4. **Basic Site Integration**
   - Integrate optimizer with existing battery control
   - Add coordination with EV charging plans
   - Implement power limit handling

### Phase 2: API & Status System
1. **API Endpoints**
   - Configuration GET/PUT endpoints
   - Status and decision explanation APIs
   - Real-time optimization state API

2. **Status Monitoring**
   - Consumption estimation and tracking
   - SoC range violation detection
   - Performance metrics calculation

3. **Error Handling & Fallbacks**
   - Rate data failure handling
   - Solar forecast integration robustness
   - Emergency charging protocols

### Phase 3: UI Implementation
1. **Simple Configuration UI**
   - On/off toggle
   - Minimum savings threshold setting
   - SoC range configuration
   - Battery wear cost setting

2. **Status Display**
   - Optimization decision explanation
   - SoC health range visualization
   - Cost savings tracking
   - Price pattern display

3. **Integration with Existing UI**
   - Add battery optimization section to dashboard
   - Mobile responsive design
   - Alert system for range violations

### Phase 4: Testing & Polish
1. **Comprehensive Testing**
   - Unit tests for all optimization algorithms
   - Integration tests with existing EVCC systems
   - Edge case validation (negative prices, pattern failures)
   - Performance testing with large rate datasets

2. **Documentation & Final Polish**
   - User configuration guide
   - Algorithm explanation documentation
   - UI/UX refinements
   - Performance optimization

## Testing Strategy

### Unit Tests
- **Battery Planner**: Algorithm correctness, edge cases
- **Charging Strategies**: Progressive vs immediate algorithms
- **Configuration**: YAML parsing, validation
- **API Endpoints**: CRUD operations, error handling

### Integration Tests
- **Site Integration**: Coordination with existing EV plans
- **Tariff Integration**: Rate data consumption
- **Battery Control**: Mode setting and coordination
- **WebSocket**: Real-time updates

### Performance Tests
- **Plan Calculation**: Performance with large rate datasets
- **Memory Usage**: Long-running plan execution
- **API Response**: Endpoint response times

### User Acceptance Tests
- **Configuration Workflow**: Plan creation and editing
- **Status Monitoring**: Real-time plan progress
- **Cost Optimization**: Actual vs expected savings

## Risks and Mitigation

### Technical Risks
1. **Battery Control Conflicts**
   - *Risk*: Conflicts between manual control, EV plans, and battery plans
   - *Mitigation*: Clear priority hierarchy and coordination logic

2. **Performance Impact**
   - *Risk*: Plan calculation affecting system responsiveness
   - *Mitigation*: Async processing, caching, and optimization

3. **Rate Data Availability**
   - *Risk*: Tariff provider outages affecting plan generation
   - *Mitigation*: Fallback strategies and graceful degradation

### User Experience Risks
1. **Configuration Complexity**
   - *Risk*: Users finding battery planning too complex
   - *Mitigation*: Smart defaults, guided setup, and clear documentation

2. **Unexpected Behavior**
   - *Risk*: Battery charging at unexpected times
   - *Mitigation*: Clear status indicators and plan preview

## Success Metrics

### Quantitative Metrics
- **Cost Savings**: Average €/day savings compared to unoptimized battery operation
- **SoC Health**: Percentage of time spent within preferred SoC range (target: >80%)
- **Pattern Detection Accuracy**: Percentage of correctly identified price peaks
- **System Performance**: Optimization decision calculation time <5 seconds
- **User Adoption**: Percentage of users enabling dynamic optimization

### Qualitative Metrics
- **Decision Transparency**: User understanding of why optimization made specific choices
- **Configuration Simplicity**: Success rate of initial setup with minimal configuration
- **System Reliability**: Uptime and graceful handling of edge cases

## Future Enhancements

### Advanced Algorithms
- **Machine Learning**: Consumption pattern learning for better predictions
- **Multi-Objective Optimization**: Balance cost, grid stability, and battery health
- **Demand Response**: Integration with grid demand response programs

### Extended Integrations
- **Home Assistant**: Native integration with HA energy dashboard
- **Vehicle-to-Home**: Coordination with V2H capable vehicles
- **Grid Services**: Participation in grid stabilization services

### Enhanced UI
- **Mobile App**: Dedicated mobile interface
- **Predictive Analytics**: Long-term cost and savings projections
- **Smart Suggestions**: AI-powered plan recommendations

## Conclusion

The dynamic SoC optimization approach leverages EVCC's existing battery management infrastructure while adding intelligent economic decision-making. By using the simple minSoC-based control strategy, the implementation complexity is minimized while delivering significant value to users facing duck curve electricity pricing.

The market-adaptive algorithms ensure batteries are optimally managed across different seasons and price patterns while maintaining system stability and user control. The feature integrates seamlessly with existing EVCC components and provides a foundation for future enhancements in whole-home energy optimization.

## Key Architectural Decisions Summary

### **1. Mode-Based Control Strategy**
- **Decision**: Use existing EVCC `SetBatteryMode()` API with mode-driven SoC targeting
- **Rationale**: Leverages proven battery management logic, eliminates need for new APIs
- **Implementation**: Algorithm calculates optimal target SoC and sets appropriate battery mode to achieve it

### **2. Market-Adaptive Thresholds**
- **Decision**: All thresholds relative to actual price data, no static price limits
- **Rationale**: Prevents breakage when markets shift (inflation, new dynamics)
- **Implementation**: Peak detection, economic viability based on price ratios and patterns

### **3. Single Economic Threshold Configuration**
- **Decision**: Single economic threshold (minSavingsPerKwh) for all charging decisions
- **Rationale**: Eliminates redundant price limits, relies on market-adaptive peak detection for economic viability
- **Implementation**: Simple UI with one clear cost threshold, maximum transparency

### **4. Battery Health Integration**
- **Decision**: SoC range management with time-based constraints
- **Rationale**: Balance economic optimization with battery longevity
- **Implementation**: Preferred range (30-80%), temporary excursions allowed, forced return after 12h

### **5. Efficiency-Aware Economics**
- **Decision**: Include round-trip efficiency (85%) and wear cost (€0.10/kWh) in all calculations
- **Rationale**: Accurate economic modeling prevents suboptimal charging decisions
- **Implementation**: True cost = (gridPrice / efficiency) + wearCost

### **6. Graceful Degradation**
- **Decision**: When rate data unavailable, defer to inverter except for emergency SoC
- **Rationale**: No price data = no economic optimization possible
- **Implementation**: Hands-off approach unless SoC drops below 15%

### **7. European Market Optimization**
- **Decision**: Leverage predictable day-ahead pricing schedule (prices known by 14:00)
- **Rationale**: Reliable 24-48h planning horizon enables confident optimization
- **Implementation**: Pattern analysis and recalculation aligned with market timing

---

## Interface Analysis & Compatibility Assessment

### **Current EVCC Battery Interfaces**

Based on analysis of the existing codebase, EVCC provides the following battery control interfaces:

#### Core Battery Interfaces (`api/api.go:113-135`)
```go
// Battery provides battery Soc in %
type Battery interface {
    Soc() (float64, error)
}

// BatteryCapacity provides a capacity in kWh
type BatteryCapacity interface {
    Capacity() float64
}

// BatteryController optionally allows to control home battery (dis)charging behavior
type BatteryController interface {
    SetBatteryMode(BatteryMode) error
}
```

#### Battery Modes (`api/batterymode.go`)
```go
type BatteryMode int

const (
    BatteryUnknown BatteryMode = iota
    BatteryNormal   // Normal operation
    BatteryHold     // Hold current SoC level
    BatteryCharge   // Force charging
)
```

#### Site-Level Control (`core/site_api.go:220-250`)
- Priority SoC thresholds (GetPrioritySoc/SetPrioritySoc)
- Buffer SoC levels (GetBufferSoc/SetBufferSoc)
- Battery mode control (GetBatteryMode/SetBatteryMode)
- Grid charge limiting (GetBatteryGridChargeLimit/SetBatteryGridChargeLimit)

### **Interface Compatibility Analysis**

#### ✅ **Excellent Compatibility - No Interface Changes Required**

The existing EVCC battery interfaces are **perfectly sufficient** for implementing the dynamic SoC optimization feature. The key insight is that the optimizer can use the existing **mode-based control strategy** instead of requiring new MinSoC control APIs.

#### **Mode-Driven SoC Optimization Strategy**

The dynamic SoC optimizer generates battery mode decisions based on current SoC vs target SoC and price patterns (detailed algorithm implementation provided in Technical Design section).

#### **Implementation Advantages**

1. **Zero Interface Changes**: Uses existing `BatteryController.SetBatteryMode()` interface
2. **Proven Integration**: Leverages existing site battery management in `core/site_battery.go:45-180`
3. **Decorator Pattern Compatibility**: Works with existing battery SoC limit decorators in `meter/battery.go`
4. **Multi-Battery Support**: Inherits existing multi-battery system coordination
5. **HTTP API Ready**: Existing `/api/batterymode/{value}` endpoints handle mode changes

#### **Optimization Workflow**

1. **Price Pattern Analysis**: Detect cheap/expensive periods and calculate optimal SoC targets
2. **Mode Decision**: Compare current vs target SoC and determine required battery mode
3. **Mode Application**: Use existing `SetBatteryMode()` to achieve SoC objectives
4. **Coordination**: Existing site logic handles EV charging conflicts and power limits

#### **Interface Utilization Mapping**

| Feature Requirement | Existing Interface | Implementation Strategy |
|---------------------|-------------------|------------------------|
| SoC Reading | `Battery.Soc()` | Direct usage for current SoC |
| Capacity Information | `BatteryCapacity.Capacity()` | Used for energy calculations |
| Charging Control | `BatteryController.SetBatteryMode()` | Mode-driven SoC targeting |
| Status Monitoring | Site-level APIs | HTTP endpoints for optimization status |
| Configuration | YAML site config | Extend existing configuration schema |

**Key Architectural Benefits:**
- **Minimal Integration Impact**: Extends existing battery control without disruption
- **Proven Reliability**: Builds on battle-tested battery management logic
- **Future-Proof**: Compatible with existing and future battery implementations
- **Simplified Testing**: Leverages existing battery control test coverage

### **Interface Assessment Conclusion**

The current EVCC battery interfaces are **fully adequate** for implementing the dynamic SoC optimization feature. The mode-based approach is actually **superior** to the originally proposed MinSoC targeting because it:

1. **Requires no new APIs** - uses existing proven interfaces
2. **Inherits all existing coordination logic** - EV charging conflicts, power limits, etc.
3. **Maintains backward compatibility** - no breaking changes
4. **Leverages decorator pattern** - works with all existing battery implementations
5. **Provides cleaner abstraction** - modes are more intuitive than raw SoC targets

**Recommendation**: Proceed with mode-based SoC optimization using existing `BatteryController.SetBatteryMode()` interface.
