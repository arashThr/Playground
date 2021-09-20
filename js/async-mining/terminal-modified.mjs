import readline from 'readline'
import crypto from 'crypto'
import { Buffer } from 'buffer'

let jobCounter = 0

async function hashLine(line) {
    const jobId = ++jobCounter
    const buf = Buffer.allocUnsafe(20)
    console.log(`\nTesting Job #${jobId}, processing ${line}`)

    for (let rounds=0; ++rounds;) {
        crypto.randomFillSync(buf)
        const hash = crypto.createHash('sha256')
            .update(line)
            .update(buf)
        const digest = hash.digest('hex')

        if (digest.match(/^0{5}/)) {
            console.log(`\tDone with job #${jobId} after ${rounds} rounds`)
            return { line, digest, rounds, buffer: buf.toString('hex') }
        }

        if (rounds % 100_000 === 0) {
            console.log(`\tSuspending job #${jobId} after ${rounds} rounds`)
            // await new Promise(setImmediate)
            console.log(`\tResuming job #${jobId} after ${rounds} rounds`)
        }
    }
}

function doIt () {
    const rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout
    })
    rl.setPrompt('> ')
    rl.on('line', async line => {
        if (line) {
            const result = await hashLine(line)
            console.log(result)
        }
        rl.prompt()
    }).on('error', err => {
        console.error('Error occurred', err)
        rl.close()
    }).prompt()
}

doIt()
