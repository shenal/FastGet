 FastGet
===========
* A cross-platform lightweight download accelerator written in GO Lang. 
* Faster downloads compared to popular programs such as wget , Curl, axel 
 
 Usage
 ------
 * specify the link with the -url flag
 * Specify the number of connections to be used with the -n flag
 * Specify the output file with the -O flag . Default would be the file name of the response
 
 Example 
 -------
 fastget -url http://abc.com/def.mp4 -n 8 -O myfile.mp4

