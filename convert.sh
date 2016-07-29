#!/bin/bash

echo "Replacing in $1..."
echo

sed -i "" "s/package\s*\(.*\)/package \1_test/" $1
sed -i "" "s/g.Assert/Expect/" $1
sed -i "" "s/g.Describe/Describe/" $1
sed -i "" "s/g.It/It/" $1
sed -i "" "s/Expect[(]err != nil[)].IsTrue[(][)]/Expect(err).To(HaveOccurred())/" $1
sed -i "" "s/Expect[(]err == nil[)].IsTrue[(][)]/Expect(err).NotTo(HaveOccurred())/" $1
sed -i "" "s@\s*[>]\s*\(.*\)[)].IsTrue[(][)]@).To(BeNumerically(\">\", \1))@" $1
sed -i "" "s@\s*[!][=]\s*\(.*\)[)].IsTrue[(][)]@).NotTo(BeEquivalentTo(\1))@" $1
sed -i "" "s/Expect.err == nil..IsFalse../Expect(err).To(HaveOccurred())/" $1
sed -i "" "s@[= ]*nil[)].IsTrue[(][)]@).To(BeNil())@" $1
sed -i "" "s@[)][.]\(Equal.*[)]\)@).To(\1)@" $1
sed -i "" "s@[)][.]\(Equal.*\)@).To(\1@" $1
sed -i "" "s@[(]int[(]\(.*\)[)][)].To[(]Equal@(\1).To(BeEquivalentTo@" $1
sed -i "" "s@IsFalse[(][)]@To(BeFalse())@" $1
sed -i "" "s@IsTrue[(][)]@To(BeTrue())@" $1
sed -i "" "s@Equal[(]nil[)]@To(BeNil())@" $1 
sed -i "" "s/To[(]To[(]/To(/" $1
