import moment from 'moment'

export const formatScore = (score) => {
  return score.toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  })
}

export const formatTimestamp = (timestamp) => {
  if (!timestamp) return ''
  return moment(timestamp * 1000).fromNow()
}

export const formatDateTime = (dateTime) => {
  if (!dateTime) return ''
  return moment(dateTime).format('MMMM D, YYYY h:mm A')
}