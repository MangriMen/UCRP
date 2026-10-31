import type { Recommendation } from './types'

const LOW_DATA = 'low_data'

export const isLowData = (recommendation: Recommendation) =>
  recommendation.factors?.some((factor) => factor.factorCode === LOW_DATA) ?? false

export const scoredFactors = (recommendation: Recommendation) =>
  (recommendation.factors ?? [])
    .filter((factor) => factor.factorCode !== LOW_DATA)
    .sort((left, right) => right.contribution - left.contribution)
