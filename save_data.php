<?php
// the $_PST[] array will contain hte passed in filename and data
// // the directory "data" is writable by the server (chmod 777)

print "hello world";

$filename = "./data/".$_POST['filename'];
$data = $_POST['filedata'];

// write the file to disk
file_put_contents($filename, $data) or die("Unable to open file!");

?>

