// 示例使用虚构数据，只用于演示转换效果
const exampleLink = 'vless://00000000-0000-4000-8000-000000000000@example.com:443?encryption=none&flow=xtls-rprx-vision&fp=chrome&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&security=reality&sid=0123456789abcdef&sni=www.example.com&spx=%2F&type=tcp#example.com'

const input = document.querySelector('#links')
const empty = document.querySelector('#empty')
const errorBox = document.querySelector('#error')
const outputBox = document.querySelector('#output')
const hint = document.querySelector('#hint')
const copyButton = document.querySelector('#copy')
const status = document.querySelector('#status')

let currentOutput = ''
let loadFailed = false
let copyTimer = 0

// 只给键名着色，通过文本节点写入，备注中的 HTML 不会被执行
function renderOutput(text) {
  for (const line of text.trimEnd().split('\n')) {
    const match = line.match(/^(\s*(?:- )?)([^\s:'"][^:]*)(:(?:\s.*)?)$/)
    if (match) {
      const key = document.createElement('span')
      key.className = 'key'
      key.textContent = match[2]
      outputBox.append(match[1], key, match[3], '\n')
    } else {
      outputBox.append(line, '\n')
    }
  }
}

function showState({ error = '', output = '' }) {
  currentOutput = output
  errorBox.textContent = error
  errorBox.hidden = !error
  outputBox.hidden = !output
  hint.hidden = !output
  copyButton.hidden = !output
  empty.hidden = Boolean(error || output)

  // 先清空再写入，报错时页面上不保留上一次的节点信息
  outputBox.replaceChildren()
  if (output) {
    renderOutput(output)
  }
}

function update() {
  if (loadFailed) {
    return
  }

  const links = input.value.trim()
  if (!links) {
    showState({})
    return
  }

  // 程序还在加载时先提示，加载完成后会重新转换
  if (!window.convertLinks) {
    showState({ error: '转换程序正在加载，请稍候。' })
    return
  }

  showState(window.convertLinks(links))
}

input.addEventListener('input', update)

document.querySelector('#example').addEventListener('click', () => {
  input.value = exampleLink
  update()
})

// 部分内嵌浏览器禁用剪贴板接口，此时退回到选中文本后复制
// 如果仍然失败，文本保持选中状态，方便手动复制
async function copyOutput() {
  try {
    await navigator.clipboard.writeText(currentOutput)
    return true
  } catch {
    const selection = getSelection()
    selection.selectAllChildren(outputBox)
    const copied = document.execCommand('copy')
    if (copied) {
      selection.removeAllRanges()
    }
    return copied
  }
}

copyButton.addEventListener('click', async () => {
  const copied = await copyOutput()
  copyButton.textContent = copied ? '已复制' : '复制失败，请手动复制'
  status.textContent = copied ? '配置已复制' : '复制失败，配置内容已选中，请手动复制'

  clearTimeout(copyTimer)
  copyTimer = setTimeout(() => {
    copyButton.textContent = '复制配置'
    status.textContent = ''
  }, 2000)
})

// 桌面端打开页面即可粘贴，手机端不主动弹出键盘
if (matchMedia('(pointer: fine)').matches) {
  input.focus()
}

const go = new Go()
WebAssembly.instantiateStreaming(fetch('v2any.wasm'), go.importObject)
  .then(({ instance }) => {
    go.run(instance)
    update()
  })
  .catch(() => {
    loadFailed = true
    showState({ error: '转换程序加载失败，请刷新页面重试。' })
  })
