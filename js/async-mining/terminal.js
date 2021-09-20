const readline = require('readline');
const crypto = require('crypto');

var job_counter=0;

async function find_hash (s, cb) {
    let hash;
    let buf = Buffer.alloc(20);
    let digest;
    let job = job_counter++;
    console.log('\n\t\tstarting job ID='+job+' processing: '+s);
    for (let rounds=0; ++rounds;) {
        crypto.randomFillSync(buf);
        hash = crypto.createHash('sha256');
        hash.update(s);
        hash.update(buf);
        digest = hash.digest('hex');
        if (digest.match(/^0{5}/)) {
            console.log('\t\tdone with job ID='+job+' after '+rounds+' rounds');
            cb(s, digest, buf.toString('hex'), rounds);
            return;
        }
        if (rounds % 100000 == 0) {
            console.log('\t\tsuspending job ID='+job+' after '+rounds+' rounds');
            await new Promise((resolve, reject) => {setImmediate(resolve)});
            console.log('\t\tresuming job ID='+job+' after '+rounds+' rounds');
        }
    }
}

function hash (line) {
    find_hash(line, (line, hash, nonce, rounds) => {
        console.log('sha('+line+'): '+hash+' nonce: '+nonce+' rounds: '+rounds);
    });
}

function doit() {
    return new Promise((resolve,reject) => {
        let rl = readline.createInterface({
            input: process.stdin,
            output: process.stdout
        });
        rl.setPrompt('> ');
        rl.on('line', (line) => {
            if (line.length) hash(line);
            rl.prompt();
        }).on('close', () => {
            resolve('all done');
        }).on('error', (err) => {
            reject(err);
        });
        rl.prompt();
    });
};

doit()
    .then((s) => console.log(s))
    .catch((e) => console.error(e));

