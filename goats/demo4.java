import java.util.Arrays;
import java.util.TreeSet;
import java.util.Random;

public class demo4 {

    /**
     * Generate a random permutation of [0...n)
     * 
     * @param n
     * @return
     */
    public static int[] permutation(int n) {
        int[] arr = new int[n];
        for (int i = 0; i < n; i++) {
            arr[i] = i;
        }

        shuffle(arr);
        return arr;
    }

    /**
     * Shuffle an integer array.
     * 
     * @param arr
     */
    public static void shuffle(int[] arr) {
        Random r = new Random();

        for (int i = arr.length - 1; i >= 0; i--) {
            int index = r.nextInt(i + 1);
            System.out.print(index + ", ");

            int tmp = arr[index];
            arr[index] = arr[i];
            arr[i] = tmp;
        }
        System.out.println();
    }

    private demo4() {
    }

     public static void main(String []args){
                int n = 52;

        int[] arr = new int[n];
        for (int i = 0; i < n; i++) {
            arr[i] = i;
        }

        //int[] arr2 = arr.clone();
        demo4.shuffle(arr);
        for (int i = arr.length - 1; i >= 0; i--) {
            System.out.print(arr[i] + ", ");
        }

        System.out.println();
     }
}