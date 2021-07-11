#!/usr/bin/env perl
use v5.26;
use warnings;

use Data::Dumper;
use Data::Printer;

use Future;
use Future::AsyncAwait;
use IO::Async::Loop;
use Syntax::Keyword::Try;

my $loop = IO::Async::Loop->new;

use Net::Async::Redis;

use constant PORT => 6309;
my $stream_name = 'mystream';

my $redis = Net::Async::Redis->new(port => PORT);
$loop->add($redis);

async sub main {
    try {
        my $res = await $redis->xread(
            COUNT => 1,
            BLOCK => 0,
            STREAMS => $stream_name, '$');
        p $res, as => 'Response: ';
    } catch ($e) {
        p $e;
    }
    main()->retain();
}

main()->retain;
$loop->run;

