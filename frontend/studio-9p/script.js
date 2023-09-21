for(const link of document.getElementsByClassName('link')) {
  console.log(link)
  link.onmousemove = (e) => {
    const decimal = e.clientX / link.offsetWidth

    const ad = 30 * decimal
    const final = 70 + ad
    console.log('------', final)

    link.style.setProperty('--light-green-percent', `${final}%`)
  }
}