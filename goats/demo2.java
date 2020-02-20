import java.util.*;

public class demo2 {
  public static void main( String args[] ) {
 	 Random seedgen = new Random();  	
 	 int seed = seedgen.nextInt();
     Random OTPGen = new Random(seed);

	 System.out.println("Nextint(10000):");
     for(int i=1; i<20; i++) {
       System.out.print(OTPGen.nextInt(10000) + ", ");
     }   
     System.out.println();

  }    
}
