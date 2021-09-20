import crypto from 'crypto'
import { Buffer } from 'buffer'
import termkit from 'terminal-kit'

const term = termkit.terminal

let colorCode = 0
let jobCounter = 0

const document = term.createDocument( {
	palette: new termkit.Palette(),
})

new termkit.Layout( {
	parent: document ,
	boxChars: 'double' ,
	layout: {
		id: 'main' ,
		y: 4,
		widthPercent: 50,
		heightPercent: 80,
		rows: [
			{
				id: 'log',
				heightPercent: 100,
			},
		]
	}
})

const logTextBox = new termkit.TextBox( {
	parent: document.elements.log,
	scrollable: true,
	vScrollBar: true,
	wordWrap: true,
	autoWidth: true,
	autoHeight: true,
    contentHasMarkup: 'ansi',
})

function getInput() {
    const input = new termkit.InlineInput( {
        parent: document,
        placeholder: 'Message to hash',
        prompt: {
            content: '> ' ,
        },
        width: 30,
    })
    input.on('submit' , value => {
        colorCode = (++colorCode) % 16
        hashLine(value, colorCode)
        document.giveFocusTo(getInput())
    })
    return input
}

async function hashLine(line, color) {
    const jobId = ++jobCounter
    const buf = Buffer.allocUnsafe(20)
    let l = term.str.green()
    logTextBox.appendLog(`\nTesting Job #${jobId}, processing ${line}`)

    for (let rounds=0; ++rounds;) {
        term.color(5)
        crypto.randomFillSync(buf)
        const hash = crypto.createHash('sha256')
            .update(line)
            .update(buf)
        const digest = hash.digest('hex')

        if (digest.match(/^0{5}/)) {
            logTextBox.appendLog(term.str.color(color, `\tDone with job #${jobId} after ${rounds} rounds`))
            return { line, digest, rounds, buffer: buf.toString('hex') }
        }

        if (rounds % 100_000 === 0) {
            logTextBox.appendLog(term.str.color(color, `\tSuspending job #${jobId} after ${rounds} rounds`))
            await new Promise(setImmediate)
            logTextBox.appendLog(term.str.color(color, `\tResuming job #${jobId} after ${rounds} rounds^`))
        }
    }
}

term.on( 'key' , key => {
	if (key === 'CTRL_C') {
        term.grabInput(false)
        term.hideCursor(false)
        term.styleReset()
        term.clear()
        term.processExit(0)
	}
})

document.giveFocusTo(getInput())
