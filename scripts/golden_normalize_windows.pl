#!/usr/bin/env perl
# Perl port of scripts/golden_normalize_windows.py, used by scripts/golden.sh
# when no working Python 3 interpreter is available.
#
# Windows ships a python.exe App Execution Alias that resolves on PATH and
# then exits with an install prompt instead of running. A `command -v` probe
# therefore reports Python as present on a machine where it cannot execute,
# which left Git Bash contributors unable to run the golden harness at all.
# Perl ships with Git for Windows, so it is always available where that shim
# is the only "Python".
#
# Behavior must stay byte-identical to the Python implementation; see its
# docstring for why each substitution exists. scripts/golden_normalize_test.sh
# pins the two against a shared fixture.

use strict;
use warnings;

if (@ARGV != 4) {
    print STDERR "usage: golden_normalize_windows.pl <actual_abs> <actual_root> <repo_root> <home>\n";
    exit 2;
}

my ($actual_abs, $actual_root, $repo_root, $home) = @ARGV;

# Backslash-form variants of a POSIX-style path, JSON-escaped form first so
# the longer prefix wins (json.dumps doubles each separator).
#
# Git Bash reports paths in MSYS form (/d/projects/x) while the generator runs
# as a native Windows binary and emits the drive form (D:\projects\x).
# Converting separators alone yields \d\projects\x, which matches neither, so
# the drive form is emitted as an additional variant -- ordered first, again
# for longest-match.
sub windows_variants {
    my ($p) = @_;
    return () unless defined $p && length $p;
    my @candidates = ($p);
    if ($p =~ m{^/([A-Za-z])/(.*)$}) {
        push @candidates, uc($1) . ":/" . $2;
    }
    my @variants;
    for my $candidate (@candidates) {
        my $win = $candidate;
        $win =~ s{/}{\\}g;
        my $win_json = $win;
        $win_json =~ s{\\}{\\\\}g;
        push @variants, $win_json, $win;
    }
    return @variants == 4 ? (@variants[2, 3], @variants[0, 1]) : @variants;
}

my $text = do { local $/; <STDIN> };
$text = '' unless defined $text;

for my $pair ([$actual_abs, '<ARTIFACT_DIR>'], [$actual_root, '<ARTIFACT_DIR>'],
              [$repo_root, '<REPO>'], [$home, '<HOME>']) {
    my ($path, $token) = @$pair;
    for my $variant (windows_variants($path)) {
        $text =~ s/\Q$variant\E/$token/g;
    }
}

# Collapse a run of backslashes to a single forward slash. JSON-string
# literals encode each separator as two backslash chars; a naive per-char
# replace would emit "//" where Linux CI compares against single-slash tokens.
$text =~ s{(<(?:ARTIFACT_DIR|REPO|HOME)>)((?:\\+[A-Za-z0-9._\-]+)+)}{
    my ($token, $rest) = ($1, $2);
    $rest =~ s{\\+}{/}g;
    $token . $rest;
}ge;

$text =~ s{(<(?:ARTIFACT_DIR|REPO)>/[A-Za-z0-9._\-/]+)\.exe}{$1}g;

print $text;
exit 0;
