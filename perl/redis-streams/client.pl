#!/usr/bin/env perl
use v5.26;
use warnings;


use Data::Dumper::Concise;
use Data::Printer;

use constant PORT => 6309;
my $stream_name = 'mystream';

my $loop = IO::Async::Loop->new;

use Net::Async::Redis;

my $redis = Net::Async::Redis->new(port => PORT);
$loop->add($redis);

my $item = 0;
async sub client {
    try {
        say "Adding $item";
        my $res = await $redis->xadd(
            $stream_name,
            MAXLEN => '=' => '2',
            '*',
            item => $item++
        );
        p $res;
    } catch ($e) {
        p $e, as => 'Error';
    }
    sleep 1;
    client()->retain();
}

client->retain;
$loop->run();

