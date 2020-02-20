import java.util.*;

public class demo1 {
  public static void main( String args[] ) {
	 Random seedgen = new Random();		
	 int seed = seedgen.nextInt();
	 System.out.println("Actual seed: " + seed);
	 Random OTPGen = new Random(seed);

	 System.out.println("Nextint:");
	 for(int i=0; i<5; i++) {
	   System.out.println(OTPGen.nextInt());
	 }	 
  }
}