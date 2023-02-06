{ 
	error[$3] += 1
} END { 
	for (e in error)
	printf "%d %d\n",e,error[e]
}