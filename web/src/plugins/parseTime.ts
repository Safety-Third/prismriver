import Vue from 'vue'

const parseTime = (time: number, recur = false): string => {
  time = Math.floor(time)
  let timeString = ''
  timeString = time % 60 < 10 ? '0' + time % 60 + timeString : time % 60 + timeString
  if (time / 60 < 1 && !recur) {
    return '0:' + timeString
  } else if (time / 60 < 1) {
    return timeString
  } else {
    return parseTime(time / 60, true) + ':' + timeString
  }
}

Vue.prototype.$parseTime = parseTime
