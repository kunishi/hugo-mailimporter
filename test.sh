#!/bin/sh

DBJAPAN_DIR=$HOME/git/dbjapan

for year in 2002 2003 2004 2005 2006 2007 2008 2009 2010 2011 2012 2013 2014; do
  rm -f result$year.txt
  for file in `cat $DBJAPAN_DIR/$year.txt`; do
    echo $file | tee -a result$year.txt
    ./hugo-mailimporter < $DBJAPAN_DIR/$file >> result$year.txt
  done
done

rm -f result.txt
for file in `ls $DBJAPAN_DIR/06_Dbjapan_mlmmj`; do
  echo $file | tee -a result.txt
  ./hugo-mailimporter < $DBJAPAN_DIR/06_Dbjapan_mlmmj/$file >> result.txt
done
