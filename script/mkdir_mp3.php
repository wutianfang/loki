<?php

$char = 'a';

$basepath = '/Users/wutianfang/go/src/github.com/wutianfang/loki/store/word_mp3/';
for($i=0; $i<26; $i++) {
    $subchar = 'a';
    for($k=0; $k<26; $k++) {
        echo $char.$subchar."\n";
        mkdir($basepath.'am/'.$char.$subchar, 0755, true);
        mkdir($basepath.'en/'.$char.$subchar, 0755, true);
        $subchar++;
    }
    $char++;

}

